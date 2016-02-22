package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mislav/go-utils/api"
	"github.com/mislav/go-utils/cli"
	"github.com/mislav/go-utils/utils"
)

var ignoredHeaders = []string{
	"access-control-allow-credentials",
	"access-control-allow-origin",
	"access-control-expose-headers",
	"connection",
	"date",
	"server",
	"status",
	"strict-transport-security",
	"via",
}

var Travis *Client

func init() {
	var logger *os.File
	if os.Getenv("TRAVIS_DEBUG") != "" {
		logger = os.Stderr
	}

	tmpdir := utils.TempDir().Join("travis")
	Travis = NewClient(logger, tmpdir.String())
}

type Client struct {
	cacheDir string
	manifest *Manifest
	http     *api.Client
}

func NewClient(logger *os.File, cacheDir string) *Client {
	rootUrl, _ := url.Parse("https://api.travis-ci.org")

	http := api.NewClient(rootUrl, func(t *api.Transport) {
		if logger != nil {
			debugStream := cli.NewColoredWriter(logger)
			debugStream.PushColor("magenta")

			t.RequestCallback = func(req *http.Request) {
				debugStream.Cprintf("> %s %C(bold)%s://%s%s%C(reset)\n", req.Method, req.URL.Scheme, req.Host, req.URL.RequestURI())
			}

			t.ResponseCallback = func(res *http.Response) {
				debugStream.Cprintf("< %C(bold)HTTP %d%C(reset)\n", res.StatusCode)

				for name, values := range res.Header {
					if ignoreHeader(name) {
						continue
					}
					value := strings.Join(values, ",")
					fmt.Fprintf(debugStream, "< %s: %s\n", name, value)
				}
			}
		}
	})

	return &Client{
		http:     http,
		cacheDir: cacheDir,
	}
}

func (c *Client) PerformRequest(method, path string, body io.Reader, configure func(*http.Request)) (*Response, error) {
	res, err := c.http.PerformRequest(method, path, nil, func(req *http.Request) {
		req.Header.Set("Travis-API-Version", "3")
		if configure != nil {
			configure(req)
		}
	})

	if err == nil {
		return &Response{Response: res}, nil
	} else {
		return nil, err
	}
}

func (c *Client) PerformAction(resourceName, actionName string, params map[string]string) (*Response, error) {
	manifest, _ := c.Manifest()
	resource := manifest.Resource(resourceName)
	if resource == nil {
		return nil, fmt.Errorf("could not find %q resource", resourceName)
	}

	var path string
	var method string
	var err error

	for _, action := range resource.NamedActions(actionName) {
		p, err := utils.ExpandUriTemplate(action.UriTemplate, params)
		if err == nil {
			path = p
			method = action.RequestMethod
			break
		}
	}

	if path == "" {
		return nil, err
	}

	return c.PerformRequest(method, path, nil, nil)
}

func (c *Client) Manifest() (*Manifest, error) {
	if c.manifest != nil {
		return c.manifest, nil
	}

	var res *Response
	var err error

	cache := utils.NewPathname(c.cacheDir, "manifest.json")
	if cache.Exists() {
		file, err := os.Open(cache.String())
		if err != nil {
			return nil, err
		}
		res = &Response{
			Response: &http.Response{Body: file},
		}
	} else {
		res, err = c.PerformRequest("GET", "/", nil, nil)
		if err != nil {
			return nil, err
		}

		cacheFile, err := cache.Create()
		if err != nil {
			return nil, err
		}

		res.Body = utils.ClosingTeeReader(res.Body, cacheFile)
	}

	c.manifest = &Manifest{}
	err = res.Unmarshal(c.manifest)
	if err != nil {
		return nil, err
	}

	return c.manifest, nil
}

type Response struct {
	*http.Response
}

func (r *Response) Unmarshal(dest interface{}) error {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)

	if err == nil {
		err = json.Unmarshal(data, dest)
	}

	return err
}

type Manifest struct {
	Config    *ManifestConfig     `json:"config"`
	Resources map[string]Resource `json:"resources"`
}

func (m *Manifest) GithubScopes() []string {
	return m.Config.GithubConfig.Scopes
}

func (m *Manifest) Resource(target string) *Resource {
	for name, resource := range m.Resources {
		if name == target {
			resource.Name = name
			return &resource
		}
	}
	return nil
}

type ManifestConfig struct {
	GithubConfig *GithubConfig `json:"github"`
}

type GithubConfig struct {
	Scopes []string `json:"scopes"`
}

type Resource struct {
	Name string
	// Actions map[string][]ResourceAction `json:"actions"`
	Actions    map[string]interface{} `json:"actions"`
	Attributes []string               `json:"attributes"`
	SortableBy []string               `json:"sortable_by"`
}

func (r *Resource) NamedActions(name string) []ResourceAction {
	result := []ResourceAction{}
	for name, actions := range r.Actions {
		switch a := actions.(type) {
		case []interface{}:
			for _, action := range a {
				action := action.(map[string]interface{})
				method := action["request_method"].(string)
				template := action["uri_template"].(string)
				result = append(result, ResourceAction{
					Name:          name,
					RequestMethod: method,
					UriTemplate:   template,
				})
			}
		}
	}
	return result
}

type ResourceAction struct {
	Name          string
	RequestMethod string `json:"request_method"`
	UriTemplate   string `json:"uri_template"`
}

func ignoreHeader(name string) bool {
	name = strings.ToLower(name)
	for _, ignored := range ignoredHeaders {
		if name == ignored {
			return true
		}
	}
	return false
}

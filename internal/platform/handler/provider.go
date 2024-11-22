package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bool64/ctxd"
	grpcRest "github.com/dohernandez/kit-template/pkg/grpc/rest"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/resources/swagger"
	mux "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v3 "github.com/swaggest/swgui/v3"
	"google.golang.org/protobuf/proto"
)

// Provider defines service locator interface.
type Provider struct {
	// Handlers contains non-api handlers to add to rest service.
	Handlers         []grpcRest.HandlerPathOption
	ResponseModifier func(context.Context, http.ResponseWriter, proto.Message) error
}

// AppendStandardHandlers registers non-api handlers.
func AppendStandardHandlers(serviceName string, p *Provider) {
	appendRootHandlers(serviceName, p)
	appendAPIDocsHandlers(serviceName, p)
}

// appendRootHandlers registers root handlers.
func appendRootHandlers(serviceName string, p *Provider) {
	// handler root path
	p.Handlers = append(p.Handlers,
		grpcRest.HandlerPathOption{
			Method:      http.MethodGet,
			PathPattern: "/",
			Handler: func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				w.Header().Set("content-type", "text/html")

				_, err := w.Write([]byte("Welcome to " + serviceName +
					`. Please read API <a href="docs">documentation</a>.`))
				must.NotFail(ctxd.WrapError(context.Background(), err, "failed to write response"))
			},
		},
	)
}

// appendAPIDocsHandlers registers handlers to display api documentation.
func appendAPIDocsHandlers(serviceName string, p *Provider) {
	// handler root path
	swh := v3.NewHandler(serviceName, "/docs/service.swagger.json", "/docs/")

	p.Handlers = append(p.Handlers,
		grpcRest.HandlerPathOption{
			Method:      http.MethodGet,
			PathPattern: "/docs/service.swagger.json",
			Handler: func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				w.Header().Set("Content-Type", "application/json")

				_, err := w.Write(swagger.SwgJSON)
				must.NotFail(ctxd.WrapError(r.Context(), err, "failed to load /docs/service.swagger.json file"))
			},
		},

		grpcRest.HandlerPathOption{
			Method:      http.MethodGet,
			PathPattern: "/docs",
			Handler: func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			},
		},

		grpcRest.HandlerPathOption{
			Method:      http.MethodGet,
			PathPattern: "/docs/swagger-ui-bundle.js",
			Handler: func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			},
		},

		grpcRest.HandlerPathOption{
			Method:      http.MethodGet,
			PathPattern: "/docs/swagger-ui-standalone-preset.js",
			Handler: func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			},
		},

		grpcRest.HandlerPathOption{
			Method:      http.MethodGet,
			PathPattern: "/docs/swagger-ui.css",
			Handler: func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				swh.ServeHTTP(w, r)
			},
		},
	)
}

// SetResponseModifier used to modify the Response status code using x-http-code header by setting a different code than 200 on success or 500 on failure.
func SetResponseModifier(p *Provider) {
	p.ResponseModifier = func(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
		md, ok := mux.ServerMetadataFromContext(ctx)
		if !ok {
			return nil
		}

		// set http status code
		if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
			code, err := strconv.Atoi(vals[0])
			if err != nil {
				return err
			}

			// delete the headers to not expose any grpc-metadata in http response
			delete(md.HeaderMD, "x-http-code")
			delete(w.Header(), "Grpc-Metadata-X-Http-Code")

			w.WriteHeader(code)
		}

		return nil
	}
}

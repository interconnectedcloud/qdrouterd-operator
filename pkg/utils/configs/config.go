package configs

import (
	"bytes"
	"text/template"

	v1alpha1 "github.com/interconnectedcloud/qdrouterd-operator/pkg/apis/interconnectedcloud/v1alpha1"
	"github.com/interconnectedcloud/qdrouterd-operator/pkg/constants"
)

func isDefaultSslProfileDefined(m *v1alpha1.Qdrouterd) bool {
	for _, profile := range m.Spec.SslProfiles {
		if profile.Name == "default" {
			return true
		}
	}
	return false
}

func isDefaultSslProfileUsed(m *v1alpha1.Qdrouterd) bool {
	for _, listener := range m.Spec.Listeners {
		if listener.SslProfile == "default" {
			return true
		}
	}
	for _, listener := range m.Spec.InterRouterListeners {
		if listener.SslProfile == "default" {
			return true
		}
	}
	return false
}

func getExposedListeners(listeners []v1alpha1.Listener) []v1alpha1.Listener {
	exposedListeners := []v1alpha1.Listener{}
	for _, listener := range listeners {
		if listener.Expose {
			exposedListeners = append(exposedListeners, listener)
		}
	}
	return exposedListeners
}

func GetQdrouterdExposedListeners(m *v1alpha1.Qdrouterd) []v1alpha1.Listener {
	listeners := []v1alpha1.Listener{}
	normal := getExposedListeners(m.Spec.Listeners)
	internal := getExposedListeners(m.Spec.InterRouterListeners)
	edge := getExposedListeners(m.Spec.EdgeListeners)
	listeners = append(listeners, normal...)
	listeners = append(listeners, internal...)
	listeners = append(listeners, edge...)
	return listeners
}

func SetQdrouterdDefaults(m *v1alpha1.Qdrouterd) bool {
	requestCert := false
	if len(m.Spec.Listeners) == 0 {
		m.Spec.Listeners = append(m.Spec.Listeners, v1alpha1.Listener{
			Port: 5672,
		}, v1alpha1.Listener{
			Port: 5671,
		}, v1alpha1.Listener{
			Port: constants.HttpLivenessPort,
			Http: true,
		})
	}
	if m.Spec.DeploymentPlan.Role == v1alpha1.RouterRoleInterior {
		if len(m.Spec.InterRouterListeners) == 0 {
			m.Spec.InterRouterListeners = append(m.Spec.InterRouterListeners, v1alpha1.Listener{
				Port: 55672,
			})
		}
		if len(m.Spec.EdgeListeners) == 0 {
			m.Spec.EdgeListeners = append(m.Spec.EdgeListeners, v1alpha1.Listener{
				Port: 45672,
			})
		}
	}
	if !isDefaultSslProfileDefined(m) && isDefaultSslProfileUsed(m) {
		m.Spec.SslProfiles = append(m.Spec.SslProfiles, v1alpha1.SslProfile{
			Name: "default",
		})
		requestCert = true
	}
	for i := range m.Spec.SslProfiles {
		if m.Spec.SslProfiles[i].Credentials == "" {
			requestCert = true
		} else if m.Spec.SslProfiles[i].RequireClientCerts && m.Spec.SslProfiles[i].CaCert == "" {
			requestCert = true
		}
	}
	return requestCert
}

func ConfigForQdrouterd(m *v1alpha1.Qdrouterd) string {
	config := `
router {
    mode: {{.DeploymentPlan.Role}}
    id: Router.${HOSTNAME}
}
{{range .Listeners}}
listener {
    {{- if .Name}}
    name: {{.Name}}
    {{- end}}
    {{- if .Host}}
    host: {{.Host}}
    {{- else}}
    host: 0.0.0.0
    {{- end}}
    {{- if .Port}}
    port: {{.Port}}
    {{- end}}
    {{- if .RouteContainer}}
    role: route-container
    {{- else }}
    role: normal
    {{- end}}
    {{- if .Http}}
    http: true
    {{- end}}
    {{- if .SslProfile}}
    sslProfile: {{.SslProfile}}
    {{- end}}
}
{{- end}}
{{range .InterRouterListeners}}
listener {
    {{- if .Name}}
    name: {{.Name}}
    {{- end}}
    role: inter-router
    {{- if .Host}}
    host: {{.Host}}
    {{- else}}
    host: 0.0.0.0
    {{- end}}
    {{- if .Port}}
    port: {{.Port}}
    {{- end}}
    {{- if .Cost}}
    cost: {{.Cost}}
    {{- end}}
    {{- if .SslProfile}}
    sslProfile: {{.SslProfile}}
    {{- end}}
}
{{- end}}
{{range .EdgeListeners}}
listener {
    {{- if .Name}}
    name: {{.Name}}
    {{- end}}
    role: edge
    {{- if .Host}}
    host: {{.Host}}
    {{- else}}
    host: 0.0.0.0
    {{- end}}
    {{- if .Port}}
    port: {{.Port}}
    {{- end}}
    {{- if .Cost}}
    cost: {{.Cost}}
    {{- end}}
    {{- if .SslProfile}}
    sslProfile: {{.SslProfile}}
    {{- end}}
}
{{- end}}
{{range .SslProfiles}}
sslProfile {
   name: {{.Name}}
   {{- if .Credentials}}
   certFile: /etc/qpid-dispatch-certs/{{.Name}}/{{.Credentials}}/tls.crt
   privateKeyFile: /etc/qpid-dispatch-certs/{{.Name}}/{{.Credentials}}/tls.key
   {{- end}}
   {{- if .CaCert}}
   caCertFile: /etc/qpid-dispatch-certs/{{.Name}}/{{.CaCert}}/ca.crt
   {{- else if .RequireClientCerts}}
   caCertFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
   {{- end}}
}
{{- end}}
{{range .Addresses}}
address {
    {{- if .Prefix}}
    prefix: {{.Prefix}}
    {{- end}}
    {{- if .Pattern}}
    pattern: {{.Pattern}}
    {{- end}}
    {{- if .Distribution}}
    distribution: {{.Distribution}}
    {{- end}}
    {{- if .Waypoint}}
    waypoint: {{.Waypoint}}
    {{- end}}
    {{- if .IngressPhase}}
    ingressPhase: {{.IngressPhase}}
    {{- end}}
    {{- if .EgressPhase}}
    egressPhase: {{.EgressPhase}}
    {{- end}}
}
{{- end}}
{{range .AutoLinks}}
autoLink {
    {{- if .Address}}
    addr: {{.Address}}
    {{- end}}
    {{- if .Direction}}
    direction: {{.Direction}}
    {{- end}}
    {{- if .ContainerId}}
    containerId: {{.ContainerId}}
    {{- end}}
    {{- if .Connection}}
    connection: {{.Connection}}
    {{- end}}
    {{- if .ExternalPrefix}}
    externalPrefix: {{.ExternalPrefix}}
    {{- end}}
    {{- if .Phase}}
    Phase: {{.Phase}}
    {{- end}}
}
{{- end}}
{{range .LinkRoutes}}
linkRoute {
    {{- if .Prefix}}
    prefix: {{.Prefix}}
    {{- end}}
    {{- if .Pattern}}
    pattern: {{.Pattern}}
    {{- end}}
    {{- if .Direction}}
    direction: {{.Direction}}
    {{- end}}
    {{- if .Connection}}
    connection: {{.Connection}}
    {{- end}}
    {{- if .ContainerId}}
    containerId: {{.ContainerId}}
    {{- end}}
    {{- if .AddExternalPrefix}}
    addExternalPrefix: {{.AddExternalPrefix}}
    {{- end}}
    {{- if .RemoveExternalPrefix}}
    removeExternalPrefix: {{.RemoveExternalPrefix}}
    {{- end}}
}
{{- end}}
{{range .Connectors}}
connector {
    {{- if .Name}}
    name: {{.Name}}
    {{- end}}
    {{- if .Host}}
    host: {{.Host}}
    {{- end}}
    {{- if .Port}}
    port: {{.Port}}
    {{- end}}
    {{- if .RouteContainer}}
    role: route-container
    {{- else}}
    role: normal
    {{- end}}
    {{- if .Cost}}
    cost: {{.Cost}}
    {{- end}}
    {{- if .SslProfile}}
    sslProfile: {{.SslProfile}}
    {{- end}}
}
{{- end}}
{{range .InterRouterConnectors}}
connector {
    {{- if .Name}}
    name: {{.Name}}
    {{- end}}
    role: inter-router
    {{- if .Host}}
    host: {{.Host}}
    {{- end}}
    {{- if .Port}}
    port: {{.Port}}
    {{- end}}
    {{- if .Cost}}
    cost: {{.Cost}}
    {{- end}}
    {{- if .SslProfile}}
    sslProfile: {{.SslProfile}}
    {{- end}}
}
{{- end}}
{{range .EdgeConnectors}}
connector {
    {{- if .Name}}
    name: {{.Name}}
    {{- end}}
    role: edge
    {{- if .Host}}
    host: {{.Host}}
    {{- end}}
    {{- if .Port}}
    port: {{.Port}}
    {{- end}}
    {{- if .Cost}}
    cost: {{.Cost}}
    {{- end}}
    {{- if .SslProfile}}
    sslProfile: {{.SslProfile}}
    {{- end}}
}
{{- end}}`
	var buff bytes.Buffer
	qdrouterdconfig := template.Must(template.New("qdrouterdconfig").Parse(config))
	qdrouterdconfig.Execute(&buff, m.Spec)
	return buff.String()
}

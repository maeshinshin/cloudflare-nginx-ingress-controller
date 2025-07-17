{{/*
Expand the name of the chart.
*/}}
{{- define "chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create ingressClass name.
*/}}
{{- define "integrated-ingress.ingressClass" -}}
{{- default "integrated-ingress" .Values.ingressClass.name }}
{{- end -}}


{{/*
Common labels
*/}}
{{- define "chart.labels" -}}
helm.sh/chart: {{ include "chart.chart" . }}
{{ include "chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the full name for the NGINX ingress-controller subchart.
This respects the fullnameOverride set in the nginxingress values.
*/}}
{{- define "integrated-ingress.nginxFullname" -}}
{{- if .Values.nginxingress.fullnameOverride -}}
{{- .Values.nginxingress.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name "nginxingress" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}


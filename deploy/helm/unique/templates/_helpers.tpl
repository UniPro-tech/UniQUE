{{- define "unique.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "unique.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- include "unique.name" . }}
{{- end }}
{{- end }}

{{- define "unique.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "unique.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "unique.serviceAccountName" -}}
{{- include "unique.fullname" . }}-{{ .component }}
{{- end }}

{{- define "unique.imageTag" -}}
{{- default .Chart.AppVersion .tag }}
{{- end }}

{{- define "unique.authKeysClaim" -}}
{{- default "auth-keys-pvc" .Values.persistence.authKeys.existingClaim }}
{{- end }}

{{- define "unique.frontKeysClaim" -}}
{{- default "front-keys-pvc" .Values.persistence.frontKeys.existingClaim }}
{{- end }}

{{- define "unique.mysqlClaim" -}}
{{- default "mysql-pvc" .Values.database.persistence.existingClaim }}
{{- end }}

{{- define "unique.migrationContainer" -}}
image: "{{ .Values.migration.image.repository }}:{{ include "unique.imageTag" (dict "tag" .Values.migration.image.tag "Chart" .Chart) }}"
imagePullPolicy: {{ .Values.migration.image.pullPolicy }}
env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: {{ .Values.database.existingSecret }}
        key: MYSQL_ACCESS_URI
{{- end }}

{{- define "unique.waitForMigration" -}}
initContainers:
  - name: wait-for-migration
    {{- include "unique.migrationContainer" . | nindent 4 }}
    args: ["wait"]
{{- end }}

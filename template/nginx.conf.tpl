server {
    {{- range .ListenPorts }}
    {{- if and $.SSLCert $.SSLKey }}
    listen {{.}} ssl;
    {{- else }}
    listen {{.}};
    {{- end }}
    {{- end }}

    server_name {{ .ServerName }};

    {{- if and .SSLCert .SSLKey }}
    ssl_certificate     {{ .SSLCert }};
    ssl_certificate_key {{ .SSLKey }};

    # SSL 配置优化
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    {{- end }}

    {{- range .Locations }}
    location {{ .Path }} {
        {{- $hasConditions := false }}
        {{- range .Upstreams }}
        {{- if or (not (isDefaultRoute .ConditionIP)) (hasHeaderCondition .Headers) }}
        {{- $hasConditions = true }}
        {{- end }}
        {{- end }}
        
        {{- if $hasConditions }}
        # 存在路由条件，使用条件判断进行路由
        set $backend "";
        {{- range $i, $upstream := .Upstreams }}
        {{- if or (not (isDefaultRoute $upstream.ConditionIP)) (hasHeaderCondition $upstream.Headers) }}
        
        # 路由规则 {{ $i }}: {{ $upstream.Target }}
        {{- if and (not (isDefaultRoute $upstream.ConditionIP)) (hasHeaderCondition $upstream.Headers) }}
        # IP + 头部条件组合
        set $match{{ $i }}_ip 0;
        set $match{{ $i }}_header 0;
        {{ generateIPCondition $upstream.ConditionIP }} {
            set $match{{ $i }}_ip 1;
        }
        {{- range $k, $v := $upstream.Headers }}
        if ($http_{{ headerToNginxVar $k }} = "{{ $v }}") {
            set $match{{ $i }}_header 1;
        }
        {{- end }}
        if ($match{{ $i }}_ip$match{{ $i }}_header = "11") {
            set $backend "{{ $upstream.Target }}";
        }
        {{- else if not (isDefaultRoute $upstream.ConditionIP) }}
        # 仅 IP 条件
        {{ generateIPCondition $upstream.ConditionIP }} {
            set $backend "{{ $upstream.Target }}";
        }
        {{- else if hasHeaderCondition $upstream.Headers }}
        # 仅头部条件
        {{- range $k, $v := $upstream.Headers }}
        if ($http_{{ headerToNginxVar $k }} = "{{ $v }}") {
            set $backend "{{ $upstream.Target }}";
        }
        {{- end }}
        {{- end }}
        {{- end }}
        {{- end }}
        
        # 默认后端处理
        {{- range .Upstreams }}
        {{- if and (isDefaultRoute .ConditionIP) (not (hasHeaderCondition .Headers)) }}
        if ($backend = "") {
            set $backend "{{ .Target }}";
        }
        {{- end }}
        {{- end }}

        set $upstream $backend;
        proxy_pass $upstream;
        {{- else }}
        # 单个上游服务器，无路由条件
        set $upstream "{{ (index .Upstreams 0).Target }}";
        proxy_pass $upstream;
        {{- end }}

        # 代理头设置
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $server_name;

        # 代理超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 缓冲设置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;

        # 错误处理
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
        proxy_next_upstream_tries 3;
        proxy_next_upstream_timeout 30s;
    }
    {{- end }}

    # 错误页面
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }

    # 安全头
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    {{- if and .SSLCert .SSLKey }}
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    {{- end }}
}
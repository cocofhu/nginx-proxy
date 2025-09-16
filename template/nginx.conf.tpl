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

    location {{ .Path }} {
        {{- if gt (len .Upstreams) 1 }}
        # 多个上游服务器，使用条件判断进行路由
        set $backend "";
        {{- range .Upstreams }}
        {{- if or (not (isDefaultRoute .ConditionIP)) (hasHeaderCondition .Headers) }}
        # 路由条件: IP={{ .ConditionIP }}{{if .Headers}}, Headers={{range $k, $v := .Headers}} {{$k}}={{$v}}{{end}}{{end}}
        {{- if and (not (isDefaultRoute .ConditionIP)) (hasHeaderCondition .Headers) }}
        # IP 和 头部条件都存在
        {{ generateIPCondition .ConditionIP }} {
            {{ generateHeaderCondition .Headers }} {
                set $backend "{{ .Target }}";
            }
        }
        {{- else if not (isDefaultRoute .ConditionIP) }}
        # 仅 IP 条件
        {{ generateIPCondition .ConditionIP }} {
            set $backend "{{ .Target }}";
        }
        {{- else if hasHeaderCondition .Headers }}
        # 仅头部条件
        {{ generateHeaderCondition .Headers }} {
            set $backend "{{ .Target }}";
        }
        {{- end }}
        {{- end }}
        {{- end }}
        {{- range .Upstreams }}
        {{- if and (isDefaultRoute .ConditionIP) (not (hasHeaderCondition .Headers)) }}
        # 默认后端
        if ($backend = "") {
            set $backend "{{ .Target }}";
        }
        {{- end }}
        {{- end }}

        # 动态解析后端域名
        set $upstream $backend;
        proxy_pass $upstream;
        {{- else }}
        # 单个上游服务器
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
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
        # 定义变量
        set $backend "{{ (index .Upstreams 0).Target }}";
        
        # 统一使用 Go 接口进行路由判断
        access_by_lua_block {
            local http = require "resty.http"
            local cjson = require "cjson"
            
            -- 只传递必要的请求信息，配置由 Go 服务查询
            local request_data = {
                path = ngx.var.uri,
                remote_addr = ngx.var.remote_addr,
                headers = ngx.req.get_headers(),
                server_name = ngx.var.server_name
            }
            
            -- 调用路由判断接口
            local httpc = http.new()
            local res, err = httpc:request_uri("http://127.0.0.1:8080/api/route", {
                method = "POST",
                body = cjson.encode(request_data),
                headers = {
                    ["Content-Type"] = "application/json"
                },
                timeout = 1000  -- 1秒超时
            })
            
            if res and res.status == 200 then
                local result = cjson.decode(res.body)
                if result.target then
                    ngx.var.backend = result.target
                else
                    -- 没有匹配的后端，返回 404
                    ngx.status = 404
                    ngx.say("404 Not Found")
                    ngx.exit(404)
                end
            else
                -- 路由服务调用失败，返回 502
                ngx.status = 502
                ngx.say("502 Bad Gateway - Route service unavailable")
                ngx.exit(502)
            end
        }

        set $upstream $backend;
        proxy_pass $upstream;

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
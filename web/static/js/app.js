// Toast 提示框类
class Toast {
    static show(message, type = 'info', duration = 3000) {
        const container = document.getElementById('toast-container');
        if (!container) return;

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;

        toast.innerHTML = `
            <div class="toast-content">
                <div class="toast-icon"></div>
                <div class="toast-message">${message}</div>
            </div>
            <button class="toast-close">&times;</button>
        `;

        container.appendChild(toast);

        // 显示动画
        setTimeout(() => {
            toast.classList.add('show');
        }, 100);

        // 关闭按钮事件
        const closeBtn = toast.querySelector('.toast-close');
        closeBtn.addEventListener('click', () => {
            this.hide(toast);
        });

        // 自动关闭
        if (duration > 0) {
            setTimeout(() => {
                this.hide(toast);
            }, duration);
        }

        return toast;
    }

    static hide(toast) {
        toast.classList.remove('show');
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }

    static success(message, duration = 3000) {
        return this.show(message, 'success', duration);
    }

    static error(message, duration = 5000) {
        return this.show(message, 'error', duration);
    }

    static warning(message, duration = 4000) {
        return this.show(message, 'warning', duration);
    }

    static info(message, duration = 3000) {
        return this.show(message, 'info', duration);
    }
}

// 全局变量
let currentCertificates = [];

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function () {
    initializeApp();
});

// 初始化应用
function initializeApp() {
    // 初始化导航
    initNavigation();

    // 默认显示仪表板
    showPage('dashboard');

    // 初始化模态框事件
    initModals();

    // 设置定时刷新证书列表
    setInterval(() => {
        if (document.getElementById('certificates-page').style.display !== 'none') {
            loadTencentCertificates();
        }
    }, 30000); // 每30秒刷新一次
}

// 初始化导航
function initNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', function (e) {
            e.preventDefault();
            const href = this.getAttribute('href');
            const page = href.substring(1); // 移除 # 号
            showPage(page);
        });
    });

    // 仪表板按钮事件
    const dashboardButtons = document.querySelectorAll('button[onclick*="showPage"]');
    dashboardButtons.forEach(button => {
        button.addEventListener('click', function () {
            const page = this.getAttribute('onclick').match(/'([^']+)'/)[1];
            showPage(page);
        });
    });
}

// 显示页面
function showPage(page) {
    // 隐藏所有页面
    document.querySelectorAll('.page-content').forEach(content => {
        content.classList.add('hidden');
    });

    // 显示目标页面
    const targetPage = document.getElementById(`${page}-page`);
    if (targetPage) {
        targetPage.classList.remove('hidden');
    }

    // 更新导航状态
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('bg-blue-50', 'text-blue-600');
        item.classList.add('text-gray-700');
    });

    const activeNav = document.querySelector(`a[href="#${page}"]`);
    if (activeNav) {
        activeNav.classList.remove('text-gray-700');
        activeNav.classList.add('bg-blue-50', 'text-blue-600');
    }

    // 加载页面数据
    if (page === 'certificates') {
        loadTencentCertificates();
    } else if (page === 'proxies') {
        loadProxies();
    }
}

// 初始化模态框
function initModals() {
    // 添加代理按钮
    const addProxyBtn = document.getElementById('add-proxy-btn');
    if (addProxyBtn) {
        addProxyBtn.addEventListener('click', function () {
            document.getElementById('add-proxy-modal').classList.remove('hidden');
            document.getElementById('add-proxy-modal').classList.add('flex');
            // 先加载证书数据，然后填充选择框
            loadCertificatesData().then(() => {
                loadCertificatesForSelect();
            });
        });
    }

    // 关闭添加代理模态框
    const closeModalBtn = document.getElementById('close-modal');
    const cancelAddProxyBtn = document.getElementById('cancel-add-proxy');
    if (closeModalBtn) {
        closeModalBtn.addEventListener('click', function () {
            document.getElementById('add-proxy-modal').classList.add('hidden');
            document.getElementById('add-proxy-modal').classList.remove('flex');
        });
    }
    if (cancelAddProxyBtn) {
        cancelAddProxyBtn.addEventListener('click', function () {
            document.getElementById('add-proxy-modal').classList.add('hidden');
            document.getElementById('add-proxy-modal').classList.remove('flex');
        });
    }

    // 添加代理表单提交
    const addProxyForm = document.getElementById('add-proxy-form');
    if (addProxyForm) {
        addProxyForm.addEventListener('submit', function (e) {
            e.preventDefault();
            createProxy();
        });
    }

    // SSL复选框事件
    const sslCheckbox = document.getElementById('proxy-ssl');
    if (sslCheckbox) {
        sslCheckbox.addEventListener('change', function () {
            const sslConfig = document.getElementById('ssl-config');
            if (this.checked) {
                sslConfig.classList.remove('hidden');
                // 当启用SSL时，确保证书列表是最新的
                if (currentCertificates && currentCertificates.length > 0) {
                    loadCertificatesForSelect();
                } else {
                    loadCertificatesData().then(() => {
                        loadCertificatesForSelect();
                    });
                }
            } else {
                sslConfig.classList.add('hidden');
            }
        });
    }

    // 添加分流规则按钮
    const addUpstreamBtn = document.getElementById('add-upstream');
    if (addUpstreamBtn) {
        addUpstreamBtn.addEventListener('click', function () {
            addUpstreamConfig();
        });
    }

    // 申请腾讯云证书按钮
    const applyTencentBtn = document.getElementById('apply-tencent-certificate-btn');
    if (applyTencentBtn) {
        applyTencentBtn.addEventListener('click', function () {
            document.getElementById('apply-tencent-certificate-modal').classList.remove('hidden');
            document.getElementById('apply-tencent-certificate-modal').classList.add('flex');
        });
    }

    // 关闭模态框按钮
    const closeTencentBtn = document.getElementById('close-tencent-certificate-modal');
    if (closeTencentBtn) {
        closeTencentBtn.addEventListener('click', function () {
            document.getElementById('apply-tencent-certificate-modal').classList.add('hidden');
            document.getElementById('apply-tencent-certificate-modal').classList.remove('flex');
        });
    }

    // 取消按钮
    const cancelTencentBtn = document.getElementById('cancel-apply-tencent-certificate');
    if (cancelTencentBtn) {
        cancelTencentBtn.addEventListener('click', function () {
            document.getElementById('apply-tencent-certificate-modal').classList.add('hidden');
            document.getElementById('apply-tencent-certificate-modal').classList.remove('flex');
        });
    }

    // 申请表单提交
    const applyForm = document.getElementById('apply-tencent-certificate-form');
    if (applyForm) {
        applyForm.addEventListener('submit', function (e) {
            e.preventDefault();
            applyTencentCertificate();
        });
    }

    // 上传证书按钮
    const addCertBtn = document.getElementById('add-certificate-btn');
    if (addCertBtn) {
        addCertBtn.addEventListener('click', function () {
            document.getElementById('add-certificate-modal').classList.remove('hidden');
            document.getElementById('add-certificate-modal').classList.add('flex');
        });
    }

    // 关闭上传证书模态框
    const closeCertModal = document.getElementById('close-certificate-modal');
    const cancelAddCert = document.getElementById('cancel-add-certificate');
    if (closeCertModal) {
        closeCertModal.addEventListener('click', function () {
            document.getElementById('add-certificate-modal').classList.add('hidden');
            document.getElementById('add-certificate-modal').classList.remove('flex');
        });
    }
    if (cancelAddCert) {
        cancelAddCert.addEventListener('click', function () {
            document.getElementById('add-certificate-modal').classList.add('hidden');
            document.getElementById('add-certificate-modal').classList.remove('flex');
        });
    }

    // 上传证书表单提交
    const uploadForm = document.getElementById('add-certificate-form');
    if (uploadForm) {
        uploadForm.addEventListener('submit', function (e) {
            e.preventDefault();
            console.log('Upload form submitted'); // 调试日志
            uploadCertificate();
        });
    }
}

// 加载所有证书列表（包括腾讯云和上传的证书）
function loadTencentCertificates() {
    // 使用统一的证书列表接口
    fetch('/api/certificates')
        .then(response => response.json())
        .then(data => {
            const allCertificates = data.certificates || [];
            currentCertificates = allCertificates;
            displayTencentCertificates(currentCertificates);
        })
        .catch(error => {
            console.error('Error loading certificates:', error);
            const container = document.getElementById('certificates-table');
            if (container) {
                container.innerHTML = '<tr><td colspan="6" class="text-center py-4 text-red-500">加载证书列表失败</td></tr>';
            }
        });
}

// 获取过期时间显示文本
function getExpiryDisplay(cert) {
    if (!cert.expires_at) {
        // 根据证书状态判断显示内容
        const status = getCertificateStatus(cert);
        if (status === 'issuing') {
            return '申请中';
        }
        return '未知';
    }

    const expiryDate = new Date(cert.expires_at);
    // 检查是否为有效日期
    if (isNaN(expiryDate.getTime()) || expiryDate.getFullYear() < 2000) {
        const status = getCertificateStatus(cert);
        if (status === 'issuing') {
            return '申请中';
        }
        return '未知';
    }

    return expiryDate.toLocaleDateString();
}

// 显示腾讯云证书列表
function displayTencentCertificates(certificates) {
    const container = document.getElementById('certificates-table');

    if (!container) {
        console.error('Certificate table container not found');
        return;
    }

    if (!certificates || certificates.length === 0) {
        container.innerHTML = '<tr><td colspan="6" class="text-center py-4 text-gray-500">暂无证书</td></tr>';
        return;
    }

    const html = certificates.map(cert => {
        const actualStatus = getCertificateStatus(cert);
        const statusClass = getStatusClass(actualStatus);
        const statusText = getStatusText(actualStatus);
        const expiresAt = getExpiryDisplay(cert);

        return `
            <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap">
                    <div class="flex items-center">
                        <span class="text-sm font-medium text-gray-900">${cert.name}</span>
                        <button onclick="editCertificateName('${cert.source === 'upload' ? cert.id : cert.source_id}', '${cert.name}', '${cert.source}')" 
                                class="ml-2 text-blue-600 hover:text-blue-800 text-xs">
                            <i class="fas fa-edit"></i>
                        </button>
                    </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${cert.domain}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${expiresAt}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusClass}">
                        ${statusText}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${cert.source === 'upload' ? '本地上传' : '腾讯云'}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div class="flex space-x-2">
                        ${getCertificateActions(cert)}
                    </div>
                </td>
            </tr>
        `;
    }).join('');

    container.innerHTML = html;
}

// 获取状态CSS类
function getStatusClass(status) {
    const statusMap = {
        'active': 'bg-green-100 text-green-800',
        'expired': 'bg-red-100 text-red-800',
        'expiring': 'bg-yellow-100 text-yellow-800',
        'renewing': 'bg-blue-100 text-blue-800',
        'issuing': 'bg-orange-100 text-orange-800'
    };
    return statusMap[status] || 'bg-gray-100 text-gray-800';
}

// 获取证书实际状态（考虑证书文件是否存在和过期时间）
function getCertificateStatus(cert) {
    // 如果没有证书文件路径，说明还在颁发中
    if (!cert.cert_path || cert.cert_path === '') {
        return 'issuing';
    }

    // 如果有明确的状态且不为空，优先使用
    if (cert.status && cert.status !== '') {
        // 对于腾讯云证书，如果状态是 renewing，直接返回
        if (cert.status === 'renewing') {
            return 'renewing';
        }
    }

    // 根据过期时间判断状态
    if (cert.expires_at) {
        const expiryDate = new Date(cert.expires_at);
        const now = new Date();
        const daysUntilExpiry = Math.ceil((expiryDate - now) / (1000 * 60 * 60 * 24));

        if (daysUntilExpiry < 0) {
            return 'expired';
        } else if (daysUntilExpiry <= 30) {
            return 'expiring';
        } else {
            return 'active';
        }
    }

    // 如果没有过期时间信息，返回原始状态或默认为 active
    return cert.status || 'active';
}

// 获取状态文本
function getStatusText(status) {
    const statusMap = {
        'active': '正常',
        'expired': '已过期',
        'expiring': '即将过期',
        'renewing': '续期中',
        'issuing': '颁发中'
    };
    return statusMap[status] || '未知状态';
}

// 获取证书操作按钮
function getCertificateActions(cert) {
    let actions = [];

    // 获取证书的实际状态
    const actualStatus = getCertificateStatus(cert);

    // 下载按钮 - 只有腾讯云证书才支持下载
    if (cert.source === 'tencent_cloud' && cert.cert_path && cert.cert_path !== '') {
        actions.push(`
            <button onclick="downloadTencentCertificate('${cert.source_id}')" 
                    class="text-blue-600 hover:text-blue-900 text-xs">
                <i class="fas fa-download"></i> 下载
            </button>
        `);
    }

    // 续期按钮 - 只有腾讯云证书才显示续期按钮
    if (cert.source === 'tencent_cloud') {
        if (actualStatus === 'issuing') {
            // 颁发中状态不显示续期按钮
        } else if (cert.status === 'renewing') {
            actions.push(`
                <button disabled class="text-gray-400 text-xs cursor-not-allowed">
                    <i class="fas fa-spinner fa-spin"></i> 续期中
                </button>
            `);
        } else {
            actions.push(`
                <button onclick="renewTencentCertificate('${cert.source_id}')" 
                        class="text-yellow-600 hover:text-yellow-900 text-xs">
                    <i class="fas fa-sync-alt"></i> 续期
                </button>
            `);
        }
    }

    // 检查状态按钮 - 只有腾讯云证书才有检查状态功能
    if (cert.source === 'tencent_cloud') {
        actions.push(`
            <button onclick="checkTencentCertificateStatus('${cert.source_id}')" 
                    class="text-green-600 hover:text-green-900 text-xs">
                <i class="fas fa-check"></i> 检查
            </button>
        `);
    }

    // 删除按钮
    if (cert.source === 'tencent_cloud') {
        actions.push(`
            <button onclick="deleteTencentCertificate('${cert.source_id}')" 
                    class="text-red-600 hover:text-red-900 text-xs">
                <i class="fas fa-trash"></i> 删除
            </button>
        `);
    } else {
        actions.push(`
            <button onclick="deleteUploadedCertificate('${cert.id}')" 
                    class="text-red-600 hover:text-red-900 text-xs">
                <i class="fas fa-trash"></i> 删除
            </button>
        `);
    }

    return actions.join('');
}

// 申请腾讯云证书
function applyTencentCertificate() {
    const domain = document.getElementById('tencent-domain').value.trim();
    const alias = document.getElementById('tencent-cert-alias').value.trim();
    const validateType = document.getElementById('tencent-validate-type').value;

    if (!domain) {
        Toast.warning('请输入域名');
        return;
    }

    const submitBtn = document.querySelector('#apply-tencent-certificate-form button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 申请中...';
    submitBtn.disabled = true;

    fetch('/api/certificates/tencent/apply', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            domain: domain,
            cert_alias: alias || domain,
            validate_type: validateType
        })
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('申请证书失败: ' + data.error);
            } else {
                Toast.success('证书申请成功，证书ID: ' + data.certificate_id);
                // 关闭模态框
                document.getElementById('apply-tencent-certificate-modal').classList.add('hidden');
                document.getElementById('apply-tencent-certificate-modal').classList.remove('flex');
                // 清空表单
                document.getElementById('tencent-domain').value = '';
                document.getElementById('tencent-cert-alias').value = '';
                // 重新加载证书列表
                loadTencentCertificates();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('申请证书失败');
        })
        .finally(() => {
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        });
}

// 加载代理列表
function loadProxies() {
    fetch('/api/rules')
        .then(response => response.json())
        .then(data => {
            const rules = data.rules || [];
            displayProxies(rules);
        })
        .catch(error => {
            console.error('Error loading proxies:', error);
            const container = document.getElementById('proxies-table');
            if (container) {
                container.innerHTML = '<tr><td colspan="5" class="text-center py-4 text-red-500">加载代理列表失败</td></tr>';
            }
        });
}

// 显示代理列表
function displayProxies(rules) {
    const container = document.getElementById('proxies-table');

    if (!container) {
        console.error('Proxies table container not found');
        return;
    }

    if (!rules || rules.length === 0) {
        container.innerHTML = '<tr><td colspan="5" class="text-center py-4 text-gray-500">暂无代理配置</td></tr>';
        return;
    }

    const html = rules.map(rule => {
        const hasSSL = rule.ssl_cert && rule.ssl_key;
        const statusClass = rule.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800';
        const statusText = rule.enabled ? '启用' : '禁用';

        // 获取第一个location的target作为显示
        const firstTarget = rule.locations && rule.locations.length > 0 ?
            (rule.locations[0].upstreams && rule.locations[0].upstreams.length > 0 ?
                rule.locations[0].upstreams[0].target : '未配置') : '未配置';

        return `
            <tr class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${rule.server_name}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${firstTarget}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusClass}">
                        ${statusText}
                    </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    ${hasSSL ? '<i class="fas fa-lock text-green-600"></i> 已启用' : '<i class="fas fa-unlock text-gray-400"></i> 未启用'}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div class="flex space-x-2">
                        <button onclick="editProxy('${rule.id}')" class="text-blue-600 hover:text-blue-900 text-xs">
                            <i class="fas fa-edit"></i> 编辑
                        </button>
                        <button onclick="deleteProxy('${rule.id}')" class="text-red-600 hover:text-red-900 text-xs">
                            <i class="fas fa-trash"></i> 删除
                        </button>
                    </div>
                </td>
            </tr>
        `;
    }).join('');

    container.innerHTML = html;
}

// 创建代理
function createProxy() {
    const domain = document.getElementById('proxy-domain').value.trim();
    const path = document.getElementById('proxy-path').value.trim() || '/';
    const sslEnabled = document.getElementById('proxy-ssl').checked;
    const httpRedirectElement = document.getElementById('proxy-http-redirect');
    const httpRedirect = httpRedirectElement ? httpRedirectElement.checked : true; // 默认启用HTTP重定向
    const certificateId = document.getElementById('proxy-certificate').value;

    if (!domain) {
        Toast.warning('请输入域名');
        return;
    }

    // 收集分流配置
    const upstreamConfigs = [];
    const upstreamElements = document.querySelectorAll('.upstream-config');

    upstreamElements.forEach(element => {
        const condition = element.querySelector('.upstream-condition').value.trim();
        const target = element.querySelector('.upstream-target').value.trim();

        if (!target) {
            Toast.warning('请填写目标地址');
            return;
        }

        // 收集头部条件
        const headers = {};
        const headerPairs = element.querySelectorAll('.header-pair');
        headerPairs.forEach(pair => {
            const key = pair.querySelector('.header-key').value.trim();
            const value = pair.querySelector('.header-value').value.trim();
            if (key && value) {
                headers[key] = value;
            }
        });

        upstreamConfigs.push({
            target: target,
            condition_ip: condition || '0.0.0.0/0',
            headers: headers
        });
    });

    if (upstreamConfigs.length === 0) {
        Toast.warning('请至少配置一个分流规则');
        return;
    }

    // 构建请求数据
    let listenPorts = [80]; // 默认启用HTTP 80

    if (sslEnabled) {
        listenPorts.push(443); // 启用HTTPS 443
        if (!httpRedirect) {
            // 如果不启用HTTP重定向，则只使用HTTPS端口
            listenPorts = [443];
        }
    }

    const requestData = {
        server_name: domain,
        listen_ports: listenPorts,
        locations: [{
            path: path,
            upstreams: upstreamConfigs
        }]
    };

    // 如果启用SSL且选择了证书，添加证书配置
    if (sslEnabled && certificateId) {
        // 从当前证书列表中找到对应的证书
        const selectedCert = currentCertificates.find(cert =>
            (cert.source === 'upload' ? cert.id : cert.source_id) === certificateId
        );

        if (selectedCert) {
            requestData.ssl_cert = selectedCert.cert_path;
            requestData.ssl_key = selectedCert.key_path;
        }
    }

    const submitBtn = document.querySelector('#add-proxy-form button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 创建中...';
    submitBtn.disabled = true;

    fetch('/api/rules', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('创建代理失败: ' + data.error);
            } else {
                Toast.success('代理创建成功');
                // 关闭模态框
                document.getElementById('add-proxy-modal').classList.add('hidden');
                document.getElementById('add-proxy-modal').classList.remove('flex');
                // 清空表单
                resetProxyForm();
                // 重新加载代理列表
                loadProxies();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('创建代理失败');
        })
        .finally(() => {
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        });
}

// 重置代理表单
function resetProxyForm() {
    document.getElementById('proxy-domain').value = '';
    document.getElementById('proxy-path').value = '/';
    document.getElementById('proxy-ssl').checked = false;
    document.getElementById('ssl-config').classList.add('hidden');
    document.getElementById('proxy-certificate').value = '';
    document.getElementById('proxy-http-redirect').checked = true; // 默认启用HTTP重定向

    // 重置分流配置为默认状态
    const upstreamContainer = document.getElementById('upstream-configs');
    upstreamContainer.innerHTML = `
        <div class="upstream-config border border-gray-200 rounded-md p-3">
            <div class="grid grid-cols-2 gap-2 mb-2">
                <div>
                    <label class="block text-xs text-gray-600 mb-1">来源IP (CIDR)</label>
                    <input type="text" class="upstream-condition w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="0.0.0.0/0" value="0.0.0.0/0">
                </div>
                <div>
                    <label class="block text-xs text-gray-600 mb-1">目标地址</label>
                    <input type="text" class="upstream-target w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="http://localhost:3000">
                </div>
            </div>
            <div>
                <label class="block text-xs text-gray-600 mb-1">HTTP头部路由 (可选)</label>
                <div class="upstream-headers space-y-1">
                    <div class="header-pair flex gap-2">
                        <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称 (如: X-API-Version)">
                        <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值 (如: v1)">
                        <button type="button" class="add-header-btn text-green-600 hover:text-green-800 px-2">+</button>
                    </div>
                </div>
            </div>
        </div>
    `;

    // 重新绑定事件
    bindUpstreamEvents();
}

// 添加分流配置
function addUpstreamConfig() {
    const container = document.getElementById('upstream-configs');
    const newConfig = document.createElement('div');
    newConfig.className = 'upstream-config border border-gray-200 rounded-md p-3';
    newConfig.innerHTML = `
        <div class="grid grid-cols-2 gap-2 mb-2">
            <div>
                <label class="block text-xs text-gray-600 mb-1">来源IP (CIDR)</label>
                <input type="text" class="upstream-condition w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="0.0.0.0/0" value="0.0.0.0/0">
            </div>
            <div>
                <label class="block text-xs text-gray-600 mb-1">目标地址</label>
                <input type="text" class="upstream-target w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="http://localhost:3000">
            </div>
        </div>
        <div>
            <label class="block text-xs text-gray-600 mb-1">HTTP头部路由 (可选)</label>
            <div class="upstream-headers space-y-1">
                <div class="header-pair flex gap-2">
                    <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称 (如: X-API-Version)">
                    <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值 (如: v1)">
                    <button type="button" class="add-header-btn text-green-600 hover:text-green-800 px-2">+</button>
                </div>
            </div>
        </div>
        <div class="mt-2">
            <button type="button" class="remove-upstream text-red-600 hover:text-red-800 text-xs">- 删除此规则</button>
        </div>
    `;

    container.appendChild(newConfig);
    bindUpstreamEvents();
}

// 绑定分流配置事件
function bindUpstreamEvents() {
    // 添加头部按钮事件
    document.querySelectorAll('.add-header-btn').forEach(btn => {
        btn.onclick = function () {
            const headersContainer = this.closest('.upstream-headers');
            const newHeaderPair = document.createElement('div');
            newHeaderPair.className = 'header-pair flex gap-2';
            newHeaderPair.innerHTML = `
                <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称">
                <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值">
                <button type="button" class="remove-header-btn text-red-600 hover:text-red-800 px-2">-</button>
            `;
            headersContainer.appendChild(newHeaderPair);
            bindUpstreamEvents();
        };
    });

    // 删除头部按钮事件
    document.querySelectorAll('.remove-header-btn').forEach(btn => {
        btn.onclick = function () {
            this.closest('.header-pair').remove();
        };
    });

    // 删除分流规则按钮事件
    document.querySelectorAll('.remove-upstream').forEach(btn => {
        btn.onclick = function () {
            this.closest('.upstream-config').remove();
        };
    });
}

// 加载证书数据（返回Promise）
function loadCertificatesData() {
    return fetch('/api/certificates')
        .then(response => response.json())
        .then(data => {
            const allCertificates = data.certificates || [];
            currentCertificates = allCertificates;
            return allCertificates;
        })
        .catch(error => {
            console.error('Error loading certificates:', error);
            currentCertificates = [];
            return [];
        });
}

// 加载证书选项
function loadCertificatesForSelect() {
    const select = document.getElementById('proxy-certificate');
    if (!select) {
        console.error('Certificate select element not found');
        return;
    }

    // 清空现有选项
    select.innerHTML = '<option value="">选择证书...</option>';

    console.log('Loading certificates for select, currentCertificates:', currentCertificates);

    // 使用已加载的证书列表
    if (currentCertificates && currentCertificates.length > 0) {
        let validCertCount = 0;
        currentCertificates.forEach(cert => {
            console.log('Processing certificate:', cert);
            // 只显示有效的证书（有证书文件路径的）
            if (cert.cert_path && cert.key_path && cert.cert_path !== '') {
                const option = document.createElement('option');
                option.value = cert.source === 'upload' ? cert.id : cert.source_id;
                option.textContent = `${cert.name || cert.domain} (${cert.domain})`;
                select.appendChild(option);
                validCertCount++;
                console.log('Added certificate option:', option.textContent);
            } else {
                console.log('Skipping certificate (no cert_path):', cert.name || cert.domain);
            }
        });

        if (validCertCount === 0) {
            const option = document.createElement('option');
            option.value = '';
            option.textContent = '暂无可用证书';
            option.disabled = true;
            select.appendChild(option);
        }

        console.log(`Loaded ${validCertCount} valid certificates`);
    } else {
        console.log('No certificates available');
        const option = document.createElement('option');
        option.value = '';
        option.textContent = '暂无证书，请先上传或申请证书';
        option.disabled = true;
        select.appendChild(option);
    }
}

// 删除代理
window.deleteProxy = function (ruleId) {
    if (!confirm('确定要删除此代理配置吗？')) {
        return;
    }

    fetch(`/api/rules/${ruleId}`, {
        method: 'DELETE'
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('删除代理失败: ' + data.error);
            } else {
                Toast.success('代理删除成功');
                loadProxies();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('删除代理失败');
        });
};

// 编辑代理（暂时用简单的提示，后续可以实现完整的编辑功能）
window.editProxy = function (ruleId) {
    Toast.info('编辑功能正在开发中，请先删除后重新创建');
};

// 初始化时绑定分流配置事件
document.addEventListener('DOMContentLoaded', function () {
    setTimeout(() => {
        bindUpstreamEvents();
    }, 100);
});

// 删除上传的证书
window.deleteUploadedCertificate = function (certificateId) {
    if (!confirm('确定要删除此证书吗？')) {
        return;
    }

    fetch(`/api/certificates/${certificateId}`, {
        method: 'DELETE'
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('删除证书失败: ' + data.error);
            } else {
                Toast.success('证书删除成功');
                loadTencentCertificates();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('删除证书失败');
        });
};

// 续期腾讯云证书
window.renewTencentCertificate = function (certificateId) {
    if (!confirm('确定要续期此证书吗？续期过程可能需要一些时间。')) {
        return;
    }

    // 获取按钮元素并添加加载状态
    const button = event.target.closest('button');
    const originalText = button.innerHTML;
    button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 续期中...';
    button.disabled = true;

    fetch(`/api/certificates/tencent/${certificateId}/renew`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('续期证书失败: ' + data.error);
            } else {
                Toast.success('证书续期已开始，新证书ID: ' + data.new_cert_id + '。请定期检查状态，续期完成后会自动切换。');
                loadTencentCertificates(); // 重新加载证书列表
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('续期证书失败');
        })
        .finally(() => {
            // 恢复按钮状态
            button.innerHTML = originalText;
            button.disabled = false;
        });
};

// 编辑证书名称
window.editCertificateName = function (certificateId, currentName, source = 'tencent_cloud') {
    const newName = prompt('请输入新的证书名称:', currentName);
    if (newName && newName !== currentName) {
        // 根据证书来源选择不同的API端点
        const apiUrl = source === 'upload'
            ? `/api/certificates/${certificateId}/name`
            : `/api/certificates/tencent/${certificateId}/name`;

        fetch(apiUrl, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({name: newName})
        })
            .then(response => response.json())
            .then(data => {
                if (data.error) {
                    Toast.error('更新证书名称失败: ' + data.error);
                } else {
                    Toast.success('证书名称更新成功');
                    loadTencentCertificates(); // 重新加载证书列表
                }
            })
            .catch(error => {
                console.error('Error:', error);
                Toast.error('更新证书名称失败');
            });
    }
};

// 检查腾讯云证书状态
window.checkTencentCertificateStatus = function (certificateId) {
    const button = event.target.closest('button');
    const originalText = button.innerHTML;
    button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 检查中...';
    button.disabled = true;

    fetch(`/api/certificates/tencent/${certificateId}/status`)
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('检查证书状态失败: ' + data.error);
            } else {
                let message = `证书状态: ${data.status}\n`;
                if (data.domain) {
                    message += `域名: ${data.domain}\n`;
                }
                if (data.expires_at) {
                    message += `过期时间: ${new Date(data.expires_at).toLocaleString()}\n`;
                }
                if (data.message) {
                    message += `备注: ${data.message}`;
                }
                Toast.info(message);

                // 重新加载证书列表以更新状态
                loadTencentCertificates();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('检查证书状态失败');
        })
        .finally(() => {
            button.innerHTML = originalText;
            button.disabled = false;
        });
};

// 下载腾讯云证书
window.downloadTencentCertificate = function (certificateId) {
    fetch(`/api/certificates/tencent/${certificateId}/download`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('下载证书失败: ' + data.error);
            } else {
                Toast.success('证书下载成功');
                loadTencentCertificates();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('下载证书失败');
        });
};

// 删除腾讯云证书
window.deleteTencentCertificate = function (certificateId) {
    if (!confirm('确定要删除此证书吗？此操作不可恢复。')) {
        return;
    }

    fetch(`/api/certificates/tencent/${certificateId}`, {
        method: 'DELETE'
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('删除证书失败: ' + data.error);
            } else {
                Toast.success('证书删除成功');
                loadTencentCertificates();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('删除证书失败');
        });
};

// 上传证书
function uploadCertificate() {
    console.log('uploadCertificate function called'); // 调试日志

    const name = document.getElementById('certificate-name').value.trim();
    const certFile = document.getElementById('certificate-cert').files[0];
    const keyFile = document.getElementById('certificate-key').files[0];

    console.log('Form data:', {name, certFile, keyFile}); // 调试日志

    if (!certFile || !keyFile) {
        Toast.warning('请选择证书文件和私钥文件');
        return;
    }

    const formData = new FormData();
    formData.append('cert', certFile);
    formData.append('key', keyFile);
    if (name) {
        formData.append('name', name);
    }

    const submitBtn = document.querySelector('#add-certificate-form button[type="submit"]');
    const originalText = submitBtn.innerHTML;
    submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 上传中...';
    submitBtn.disabled = true;

    fetch('/api/certificates', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                Toast.error('上传证书失败: ' + data.error);
            } else {
                Toast.success('证书上传成功');
                // 关闭模态框
                document.getElementById('add-certificate-modal').classList.add('hidden');
                document.getElementById('add-certificate-modal').classList.remove('flex');
                // 清空表单
                document.getElementById('certificate-name').value = '';
                document.getElementById('certificate-cert').value = '';
                document.getElementById('certificate-key').value = '';
                // 重新加载证书列表
                loadTencentCertificates();
            }
        })
        .catch(error => {
            console.error('Error:', error);
            Toast.error('上传证书失败: ' + error.message);
        })
        .finally(() => {
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        });
}
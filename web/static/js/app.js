// 全局变量
let currentPage = 'dashboard';

// DOM加载完成后初始化
document.addEventListener('DOMContentLoaded', function () {
    initializeApp();
});

// 初始化应用
function initializeApp() {
    setupEventListeners();
    setupNavigation();
    bindUpstreamEvents();
}

// 设置事件监听器
function setupEventListeners() {
    // 侧边栏切换
    const sidebarToggle = document.getElementById('sidebar-toggle');
    const sidebar = document.getElementById('sidebar');

    sidebarToggle?.addEventListener('click', () => {
        sidebar.classList.toggle('-translate-x-full');
    });

    // 添加代理按钮
    const addProxyBtn = document.getElementById('add-proxy-btn');
    const addProxyModal = document.getElementById('add-proxy-modal');
    const closeModal = document.getElementById('close-modal');
    const cancelAddProxy = document.getElementById('cancel-add-proxy');

    addProxyBtn?.addEventListener('click', () => {
        addProxyModal.classList.remove('hidden');
        addProxyModal.classList.add('flex');
    });

    closeModal?.addEventListener('click', closeAddProxyModal);
    cancelAddProxy?.addEventListener('click', closeAddProxyModal);

    // 添加代理表单提交
    const addProxyForm = document.getElementById('add-proxy-form');
    addProxyForm?.addEventListener('submit', handleAddProxy);

    // 添加分流规则按钮
    const addUpstreamBtn = document.getElementById('add-upstream');
    addUpstreamBtn?.addEventListener('click', addUpstreamConfig);

    // 证书管理相关事件
    const addCertificateBtn = document.getElementById('add-certificate-btn');
    const addCertificateModal = document.getElementById('add-certificate-modal');
    const closeCertificateModal = document.getElementById('close-certificate-modal');
    const cancelAddCertificate = document.getElementById('cancel-add-certificate');

    addCertificateBtn?.addEventListener('click', () => {
        addCertificateModal.classList.remove('hidden');
        addCertificateModal.classList.add('flex');
    });

    closeCertificateModal?.addEventListener('click', closeAddCertificateModal);
    cancelAddCertificate?.addEventListener('click', closeAddCertificateModal);

    // 证书上传表单提交
    const addCertificateForm = document.getElementById('add-certificate-form');
    addCertificateForm?.addEventListener('submit', handleAddCertificate);

    // SSL复选框变化事件
    const sslCheckbox = document.getElementById('proxy-ssl');
    const sslConfig = document.getElementById('ssl-config');
    sslCheckbox?.addEventListener('change', function () {
        if (this.checked) {
            sslConfig.classList.remove('hidden');
            loadCertificatesForSelect();
        } else {
            sslConfig.classList.add('hidden');
        }
    });

    // 编辑代理相关事件
    const closeEditModal = document.getElementById('close-edit-modal');
    const cancelEditProxy = document.getElementById('cancel-edit-proxy');

    closeEditModal?.addEventListener('click', closeEditProxyModal);
    cancelEditProxy?.addEventListener('click', closeEditProxyModal);

    // 编辑代理表单提交
    const editProxyForm = document.getElementById('edit-proxy-form');
    editProxyForm?.addEventListener('submit', handleEditProxy);

    // 编辑分流规则按钮
    const editAddUpstreamBtn = document.getElementById('edit-add-upstream');
    editAddUpstreamBtn?.addEventListener('click', () => addEditUpstreamConfig());

    // 编辑SSL复选框变化事件
    const editSslCheckbox = document.getElementById('edit-proxy-ssl');
    const editSslConfig = document.getElementById('edit-ssl-config');
    editSslCheckbox?.addEventListener('change', function () {
        if (this.checked) {
            editSslConfig.classList.remove('hidden');
            loadCertificatesForEditSelect();
        } else {
            editSslConfig.classList.add('hidden');
        }
    });
}

// 设置导航
function setupNavigation() {
    const navItems = document.querySelectorAll('.nav-item');

    navItems.forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const href = item.getAttribute('href');
            const page = href.substring(1);
            navigateToPage(page);
        });
    });
}

// 页面导航
function navigateToPage(page) {
    // 隐藏所有页面
    const pages = document.querySelectorAll('.page-content');
    pages.forEach(p => p.classList.add('hidden'));

    // 显示目标页面
    const targetPage = document.getElementById(`${page}-page`);
    if (targetPage) {
        targetPage.classList.remove('hidden');
        currentPage = page;

        // 更新导航状态
        updateNavigation(page);

        // 加载页面数据
        loadPageData(page);
    }
}

// showPage 函数别名，供HTML中的按钮调用
window.showPage = function (page) {
    navigateToPage(page);
};

// 更新导航状态
function updateNavigation(activePage) {
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        const href = item.getAttribute('href');
        const page = href.substring(1);

        if (page === activePage) {
            item.classList.add('bg-blue-50', 'text-blue-600');
        } else {
            item.classList.remove('bg-blue-50', 'text-blue-600');
        }
    });
}


// 加载页面数据
function loadPageData(page) {
    switch (page) {
        case 'proxies':
            loadProxiesData();
            break;
        case 'certificates':
            loadCertificatesData();
            break;
    }
}

// 加载代理数据
async function loadProxiesData() {
    const proxiesTable = document.getElementById('proxies-table');
    if (!proxiesTable) return;

    try {
        const response = await apiCall('/rules');
        const rules = response.rules || [];

        proxiesTable.innerHTML = rules.map(rule => {
            // 生成目标地址和路由条件的显示
            const targetInfo = rule.locations.map(loc => {
                return loc.upstreams.map(up => {
                    let info = up.target;
                    const conditions = [];

                    if (up.condition_ip && up.condition_ip !== '0.0.0.0/0') {
                        conditions.push(`IP: ${up.condition_ip}`);
                    }

                    if (up.headers && Object.keys(up.headers).length > 0) {
                        const headerStr = Object.entries(up.headers)
                            .map(([k, v]) => `${k}=${v}`)
                            .join(', ');
                        conditions.push(`Headers: ${headerStr}`);
                    }

                    if (conditions.length > 0) {
                        info += ` <small class="text-gray-400">(${conditions.join('; ')})</small>`;
                    }

                    return info;
                }).join('<br>');
            }).join('<br>');

            return `
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${rule.server_name}</td>
                    <td class="px-6 py-4 text-sm text-gray-500">${targetInfo}</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                            活跃
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${rule.ssl_cert ? '<i class="fas fa-lock text-green-600"></i>' : '<i class="fas fa-unlock text-red-600"></i>'}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <button onclick="editRule('${rule.id}')" class="text-indigo-600 hover:text-indigo-900 mr-3">编辑</button>
                        <button onclick="deleteRule('${rule.id}')" class="text-red-600 hover:text-red-900">删除</button>
                    </td>
                </tr>
            `;
        }).join('');

    } catch (error) {
        console.error('Failed to load proxies data:', error);
        proxiesTable.innerHTML = '<tr><td colspan="5" class="text-center py-4 text-gray-500">加载失败</td></tr>';
    }
}


// 关闭添加代理模态框
function closeAddProxyModal() {
    const modal = document.getElementById('add-proxy-modal');
    modal.classList.add('hidden');
    modal.classList.remove('flex');

    // 清空表单
    document.getElementById('add-proxy-form').reset();

    // 重置分流配置为默认状态
    const upstreamContainer = document.getElementById('upstream-configs');
    upstreamContainer.innerHTML = createDefaultUpstreamConfig();

    // 重新绑定事件
    bindUpstreamEvents();

    // 重置SSL配置
    const sslCheckbox = document.getElementById('proxy-ssl');
    const sslConfig = document.getElementById('ssl-config');
    sslCheckbox.checked = false;
    sslConfig.classList.add('hidden');
}

// 创建默认的upstream配置HTML
function createDefaultUpstreamConfig() {
    return `
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
}

// 绑定upstream相关事件
function bindUpstreamEvents() {
    // 绑定添加头部按钮事件
    document.querySelectorAll('.add-header-btn').forEach(btn => {
        btn.addEventListener('click', function () {
            addHeaderPair(this.closest('.upstream-headers'));
        });
    });
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
                <input type="text" class="upstream-condition w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="192.168.1.0/24">
            </div>
            <div>
                <label class="block text-xs text-gray-600 mb-1">目标地址</label>
                <input type="text" class="upstream-target w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="http://internal-server:8080">
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
        <button type="button" class="remove-upstream mt-2 text-xs text-red-600 hover:text-red-800">移除</button>
    `;

    // 添加移除按钮事件
    newConfig.querySelector('.remove-upstream').addEventListener('click', function () {
        newConfig.remove();
    });

    // 添加头部按钮事件
    newConfig.querySelector('.add-header-btn').addEventListener('click', function () {
        addHeaderPair(this.closest('.upstream-headers'));
    });

    container.appendChild(newConfig);
}

// 添加头部键值对
function addHeaderPair(headersContainer) {
    // 先将现有的"+"按钮改为"-"按钮
    const existingAddBtn = headersContainer.querySelector('.add-header-btn');
    if (existingAddBtn) {
        existingAddBtn.textContent = '-';
        existingAddBtn.className = 'remove-header-btn text-red-600 hover:text-red-800 px-2';
        existingAddBtn.removeEventListener('click', arguments.callee);
        existingAddBtn.addEventListener('click', function () {
            existingAddBtn.closest('.header-pair').remove();
        });
    }

    const newHeaderPair = document.createElement('div');
    newHeaderPair.className = 'header-pair flex gap-2';
    newHeaderPair.innerHTML = `
        <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称">
        <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值">
        <button type="button" class="add-header-btn text-green-600 hover:text-green-800 px-2">+</button>
    `;

    // 添加新的"+"按钮事件
    newHeaderPair.querySelector('.add-header-btn').addEventListener('click', function () {
        addHeaderPair(headersContainer);
    });

    headersContainer.appendChild(newHeaderPair);
}

// 处理添加代理
async function handleAddProxy(e) {
    e.preventDefault();

    const domain = document.getElementById('proxy-domain').value;
    const path = document.getElementById('proxy-path').value || '/';
    const ssl = document.getElementById('proxy-ssl').checked;
    const certificateId = document.getElementById('proxy-certificate').value;

    // 收集所有分流配置
    const upstreamConfigs = [];
    const upstreamElements = document.querySelectorAll('.upstream-config');

    upstreamElements.forEach(element => {
        const condition = element.querySelector('.upstream-condition').value.trim();
        const target = element.querySelector('.upstream-target').value.trim();

        if (condition && target) {
            const upstream = {
                condition_ip: condition,
                target: target
            };

            // 收集头部信息
            const headers = {};
            const headerPairs = element.querySelectorAll('.header-pair');
            headerPairs.forEach(pair => {
                const key = pair.querySelector('.header-key').value.trim();
                const value = pair.querySelector('.header-value').value.trim();
                if (key && value) {
                    headers[key] = value;
                }
            });

            if (Object.keys(headers).length > 0) {
                upstream.headers = headers;
            }

            upstreamConfigs.push(upstream);
        }
    });

    if (upstreamConfigs.length === 0) {
        showNotification('请至少配置一个分流规则', 'warning');
        return;
    }

    let sslCert = '';
    let sslKey = '';

    if (ssl && certificateId) {
        // 获取选中的证书信息
        try {
            const certResponse = await apiCall(`/certificates/${certificateId}`);
            sslCert = certResponse.cert_path;
            sslKey = certResponse.key_path;
        } catch (error) {
            showNotification('获取证书信息失败', 'error');
            return;
        }
    }

    const ruleData = {
        server_name: domain,
        listen_ports: ssl ? [443] : [80],
        ssl_cert: sslCert,
        ssl_key: sslKey,
        locations: [{
            path: path,
            upstreams: upstreamConfigs
        }]
    };

    try {
        await apiCall('/rules', 'POST', ruleData);
        showNotification('代理配置已添加', 'success');
        closeAddProxyModal();

        // 重新加载代理列表
        if (currentPage === 'proxies') {
            loadProxiesData();
        }

    } catch (error) {
        console.error('Failed to add proxy:', error);
        showNotification('添加代理失败: ' + error.message, 'error');
    }
}


// 显示通知
function showNotification(message, type = 'info') {
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 z-50 p-4 rounded-lg shadow-lg transition-all duration-300 transform translate-x-full`;

    // 根据类型设置样式
    switch (type) {
        case 'success':
            notification.classList.add('bg-green-500', 'text-white');
            break;
        case 'error':
            notification.classList.add('bg-red-500', 'text-white');
            break;
        case 'warning':
            notification.classList.add('bg-yellow-500', 'text-white');
            break;
        default:
            notification.classList.add('bg-blue-500', 'text-white');
    }

    notification.innerHTML = `
        <div class="flex items-center">
            <span>${message}</span>
            <button class="ml-4 text-white hover:text-gray-200" onclick="this.parentElement.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        </div>
    `;

    document.body.appendChild(notification);

    // 显示动画
    setTimeout(() => {
        notification.classList.remove('translate-x-full');
    }, 100);

    // 自动隐藏
    setTimeout(() => {
        notification.classList.add('translate-x-full');
        setTimeout(() => {
            notification.remove();
        }, 300);
    }, 5000);
}


// 编辑规则
window.editRule = async function (ruleId) {
    try {
        const response = await apiCall(`/rules/${ruleId}`);

        // 填充编辑表单
        await populateEditForm(response);

        // 显示编辑模态框
        showEditProxyModal();

    } catch (error) {
        console.error('Failed to load rule for editing:', error);
        showNotification('加载规则失败', 'error');
    }
}

// 显示编辑代理模态框
function showEditProxyModal() {
    const modal = document.getElementById('edit-proxy-modal');
    modal.classList.remove('hidden');
    modal.classList.add('flex');

    // 加载证书列表到选择框
    loadCertificatesForEditSelect();
}

// 关闭编辑代理模态框
function closeEditProxyModal() {
    const modal = document.getElementById('edit-proxy-modal');
    modal.classList.add('hidden');
    modal.classList.remove('flex');

    // 清空表单
    document.getElementById('edit-proxy-form').reset();
    document.getElementById('edit-upstream-configs').innerHTML = '';

    // 重置SSL配置
    const sslCheckbox = document.getElementById('edit-proxy-ssl');
    const sslConfig = document.getElementById('edit-ssl-config');
    sslCheckbox.checked = false;
    sslConfig.classList.add('hidden');
}

// 填充编辑表单
async function populateEditForm(rule) {
    // 填充基本信息
    document.getElementById('edit-proxy-id').value = rule.id;
    document.getElementById('edit-proxy-domain').value = rule.server_name;

    // 填充路径（取第一个location的路径）
    const firstLocation = rule.locations && rule.locations[0];
    if (firstLocation) {
        document.getElementById('edit-proxy-path').value = firstLocation.path || '/';
    }

    // 填充SSL配置
    const sslCheckbox = document.getElementById('edit-proxy-ssl');
    const sslConfig = document.getElementById('edit-ssl-config');

    if (rule.ssl_cert) {
        sslCheckbox.checked = true;
        sslConfig.classList.remove('hidden');

        // 根据证书路径找到对应的证书ID
        try {
            const certResponse = await apiCall('/certificates');
            const certificates = certResponse.certificates || [];
            const matchingCert = certificates.find(cert =>
                cert.cert_path === rule.ssl_cert || cert.key_path === rule.ssl_key
            );

            if (matchingCert) {
                // 延迟设置选中的证书，等待证书列表加载完成
                setTimeout(() => {
                    const certSelect = document.getElementById('edit-proxy-certificate');
                    if (certSelect) {
                        certSelect.value = matchingCert.id;
                    }
                }, 100);
            }
        } catch (error) {
            console.error('Failed to load certificates for edit form:', error);
        }
    } else {
        sslCheckbox.checked = false;
        sslConfig.classList.add('hidden');
    }

    // 填充分流配置
    const upstreamContainer = document.getElementById('edit-upstream-configs');
    upstreamContainer.innerHTML = '';

    if (firstLocation && firstLocation.upstreams) {
        firstLocation.upstreams.forEach(upstream => {
            addEditUpstreamConfig(upstream.condition_ip, upstream.target, upstream.headers || {});
        });
    } else {
        // 如果没有分流配置，添加一个默认的
        addEditUpstreamConfig('0.0.0.0/0', '', {});
    }
}

// 添加编辑分流配置
function addEditUpstreamConfig(condition = '', target = '', headers = {}) {
    const container = document.getElementById('edit-upstream-configs');
    const newConfig = document.createElement('div');
    newConfig.className = 'upstream-config border border-gray-200 rounded-md p-3';

    // 生成头部HTML
    let headersHtml = '';
    if (Object.keys(headers).length > 0) {
        headersHtml = Object.entries(headers).map(([key, value]) => `
            <div class="header-pair flex gap-2">
                <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称" value="${key}">
                <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值" value="${value}">
                <button type="button" class="remove-header-btn text-red-600 hover:text-red-800 px-2">-</button>
            </div>
        `).join('');
    }

    // 如果没有现有的headers，添加一个空的header输入行
    if (Object.keys(headers).length === 0) {
        headersHtml = `
            <div class="header-pair flex gap-2">
                <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称 (如: X-API-Version)">
                <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值 (如: v1)">
                <button type="button" class="add-header-btn text-green-600 hover:text-green-800 px-2">+</button>
            </div>
        `;
    } else {
        // 如果有现有的headers，在最后一个header后添加一个"+"按钮
        headersHtml += `
            <div class="header-pair flex gap-2">
                <input type="text" class="header-key w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header名称 (如: X-API-Version)">
                <input type="text" class="header-value w-1/2 border border-gray-300 rounded px-2 py-1 text-sm" placeholder="Header值 (如: v1)">
                <button type="button" class="add-header-btn text-green-600 hover:text-green-800 px-2">+</button>
            </div>
        `;
    }

    newConfig.innerHTML = `
        <div class="grid grid-cols-2 gap-2 mb-2">
            <div>
                <label class="block text-xs text-gray-600 mb-1">来源IP (CIDR)</label>
                <input type="text" class="upstream-condition w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="0.0.0.0/0" value="${condition}">
            </div>
            <div>
                <label class="block text-xs text-gray-600 mb-1">目标地址</label>
                <input type="text" class="upstream-target w-full border border-gray-300 rounded px-2 py-1 text-sm" placeholder="http://localhost:3000" value="${target}">
            </div>
        </div>
        <div>
            <label class="block text-xs text-gray-600 mb-1">HTTP头部路由 (可选)</label>
            <div class="upstream-headers space-y-1">
                ${headersHtml}
            </div>
        </div>
        <button type="button" class="remove-upstream mt-2 text-xs text-red-600 hover:text-red-800">移除</button>
    `;

    // 添加移除按钮事件
    newConfig.querySelector('.remove-upstream').addEventListener('click', function () {
        newConfig.remove();
    });

    // 添加头部相关事件
    const addHeaderBtn = newConfig.querySelector('.add-header-btn');
    if (addHeaderBtn) {
        addHeaderBtn.addEventListener('click', function () {
            addHeaderPair(this.closest('.upstream-headers'));
        });
    }

    newConfig.querySelectorAll('.remove-header-btn').forEach(btn => {
        btn.addEventListener('click', function () {
            btn.closest('.header-pair').remove();
        });
    });

    container.appendChild(newConfig);
}

// 加载证书到编辑选择框
async function loadCertificatesForEditSelect() {
    const certificateSelect = document.getElementById('edit-proxy-certificate');
    if (!certificateSelect) return;

    try {
        const response = await apiCall('/certificates');
        const certificates = response.certificates || [];

        certificateSelect.innerHTML = '<option value="">选择证书...</option>' +
            certificates.map(cert => `
                <option value="${cert.id}">${cert.name} (${cert.domain || 'N/A'})</option>
            `).join('');

    } catch (error) {
        console.error('Failed to load certificates for edit select:', error);
        certificateSelect.innerHTML = '<option value="">加载失败</option>';
    }
}

// 处理编辑代理提交
async function handleEditProxy(e) {
    e.preventDefault();

    const ruleId = document.getElementById('edit-proxy-id').value;
    const domain = document.getElementById('edit-proxy-domain').value;
    const path = document.getElementById('edit-proxy-path').value || '/';
    const ssl = document.getElementById('edit-proxy-ssl').checked;
    const certificateId = document.getElementById('edit-proxy-certificate').value;

    // 收集所有分流配置
    const upstreamConfigs = [];
    const upstreamElements = document.querySelectorAll('#edit-upstream-configs .upstream-config');

    upstreamElements.forEach(element => {
        const condition = element.querySelector('.upstream-condition').value.trim();
        const target = element.querySelector('.upstream-target').value.trim();

        if (condition && target) {
            const upstream = {
                condition_ip: condition,
                target: target
            };

            // 收集头部信息
            const headers = {};
            const headerPairs = element.querySelectorAll('.header-pair');
            headerPairs.forEach(pair => {
                const key = pair.querySelector('.header-key').value.trim();
                const value = pair.querySelector('.header-value').value.trim();
                if (key && value) {
                    headers[key] = value;
                }
            });

            if (Object.keys(headers).length > 0) {
                upstream.headers = headers;
            }

            upstreamConfigs.push(upstream);
        }
    });

    if (upstreamConfigs.length === 0) {
        showNotification('请至少配置一个分流规则', 'warning');
        return;
    }

    let sslCert = '';
    let sslKey = '';

    if (ssl && certificateId) {
        // 获取选中的证书信息
        try {
            const certResponse = await apiCall(`/certificates/${certificateId}`);
            sslCert = certResponse.cert_path;
            sslKey = certResponse.key_path;
        } catch (error) {
            showNotification('获取证书信息失败', 'error');
            return;
        }
    }

    const ruleData = {
        server_name: domain,
        listen_ports: ssl ? [443] : [80],
        ssl_cert: sslCert,
        ssl_key: sslKey,
        locations: [{
            path: path,
            upstreams: upstreamConfigs
        }]
    };

    try {
        await apiCall(`/rules/${ruleId}`, 'PUT', ruleData);
        showNotification('代理配置已更新', 'success');
        closeEditProxyModal();

        // 重新加载代理列表
        if (currentPage === 'proxies') {
            loadProxiesData();
        }

    } catch (error) {
        console.error('Failed to update proxy:', error);
        showNotification('更新代理失败: ' + error.message, 'error');
    }
}

// 删除规则
window.deleteRule = async function (ruleId) {
    if (!confirm('确定要删除这个代理规则吗？')) {
        return;
    }

    try {
        await apiCall(`/rules/${ruleId}`, 'DELETE');
        showNotification('代理规则已删除', 'success');

        // 重新加载代理列表
        if (currentPage === 'proxies') {
            loadProxiesData();
        }

    } catch (error) {
        console.error('Failed to delete rule:', error);
        showNotification('删除规则失败: ' + error.message, 'error');
    }
}

// 加载证书数据
async function loadCertificatesData() {
    const certificatesTable = document.getElementById('certificates-table');
    if (!certificatesTable) return;

    try {
        const response = await apiCall('/certificates');
        const certificates = response.certificates || [];

        certificatesTable.innerHTML = certificates.map(cert => {
            const expiresAt = new Date(cert.expires_at);
            const now = new Date();
            const daysUntilExpiry = Math.ceil((expiresAt - now) / (1000 * 60 * 60 * 24));

            let statusClass = 'bg-green-100 text-green-800';
            let statusText = '有效';

            if (daysUntilExpiry < 0) {
                statusClass = 'bg-red-100 text-red-800';
                statusText = '已过期';
            } else if (daysUntilExpiry < 30) {
                statusClass = 'bg-yellow-100 text-yellow-800';
                statusText = '即将过期';
            }

            return `
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${cert.name}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${cert.domain || '-'}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${expiresAt.toLocaleDateString()}</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${statusClass}">
                            ${statusText}
                        </span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <button onclick="deleteCertificate('${cert.id}')" class="text-red-600 hover:text-red-900">删除</button>
                    </td>
                </tr>
            `;
        }).join('');

    } catch (error) {
        console.error('Failed to load certificates data:', error);
        certificatesTable.innerHTML = '<tr><td colspan="5" class="text-center py-4 text-gray-500">加载失败</td></tr>';
    }
}

// 加载证书到选择框
async function loadCertificatesForSelect() {
    const certificateSelect = document.getElementById('proxy-certificate');
    if (!certificateSelect) return;

    try {
        const response = await apiCall('/certificates');
        const certificates = response.certificates || [];

        certificateSelect.innerHTML = '<option value="">选择证书...</option>' +
            certificates.map(cert => `
                <option value="${cert.id}">${cert.name} (${cert.domain || 'N/A'})</option>
            `).join('');

    } catch (error) {
        console.error('Failed to load certificates for select:', error);
        certificateSelect.innerHTML = '<option value="">加载失败</option>';
    }
}

// 关闭证书上传模态框
function closeAddCertificateModal() {
    const modal = document.getElementById('add-certificate-modal');
    modal.classList.add('hidden');
    modal.classList.remove('flex');

    // 清空表单
    document.getElementById('add-certificate-form').reset();
}

// 处理证书上传
async function handleAddCertificate(e) {
    e.preventDefault();

    const name = document.getElementById('certificate-name').value;
    const certFile = document.getElementById('certificate-cert').files[0];
    const keyFile = document.getElementById('certificate-key').files[0];

    if (!name || !certFile || !keyFile) {
        showNotification('请填写所有必需字段', 'warning');
        return;
    }

    const formData = new FormData();
    formData.append('cert', certFile);
    formData.append('key', keyFile);

    try {
        const response = await fetch('/api/certificates', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        showNotification('证书上传成功', 'success');
        closeAddCertificateModal();

        // 重新加载证书列表
        if (currentPage === 'certificates') {
            loadCertificatesData();
        }

    } catch (error) {
        console.error('Failed to upload certificate:', error);
        showNotification('证书上传失败: ' + error.message, 'error');
    }
}

// 删除证书
window.deleteCertificate = async function (certificateId) {
    if (!confirm('确定要删除这个证书吗？')) {
        return;
    }

    try {
        await apiCall(`/certificates/${certificateId}`, 'DELETE');
        showNotification('证书已删除', 'success');

        // 重新加载证书列表
        if (currentPage === 'certificates') {
            loadCertificatesData();
        }

    } catch (error) {
        console.error('Failed to delete certificate:', error);
        showNotification('删除证书失败: ' + error.message, 'error');
    }
}

// API调用函数
async function apiCall(endpoint, method = 'GET', data = null) {
    try {
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
            }
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        const response = await fetch(`/api${endpoint}`, options);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        showNotification('API调用失败', 'error');
        throw error;
    }
}
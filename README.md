# Kubernetes Dashboard 产品需求文档（PRD）

## 1. 项目背景
Kubernetes 已成为主流容器编排工具，但官方 Dashboard 功能有限，UI 较为简陋，不够优雅和直观。为了便于运维和开发人员更高效地管理集群，计划开发一个 **基于 Go 语言的前端 Dashboard**，前期不依赖真实后端服务，而是通过 **Mock 数据与 REST API** 提供交互，保证系统的 UI 与交互体验先行落地。

---

## 2. 产品目标
- 提供一个美观、直观的 Kubernetes 集群管理界面。  
- 初期通过 Mock 数据填充，支持常用的 REST API 交互。  
- 方便后期替换为真实 Kubernetes API Server 或自研后端服务。  
- 前端重点打造 **优雅的 UI/UX**，确保操作顺畅。  

---

## 3. 技术栈
- **前端框架**：Go + WebAssembly（推荐 Vugu / Gio / 结合 Vue 前端框架）  
- **UI 框架**：Tailwind CSS / Ant Design Vue  
- **Mock 数据**：内置 JSON 文件 / Mock 服务（如 `go-fakeit`）  
- **接口风格**：RESTful API  

---

## 4. 功能需求
### 4.1 集群概览
- 集群基本信息（版本、节点数、命名空间数、Pod 数量）  
- CPU / 内存使用率（Mock 数据填充）  
- 最近事件（Events）  

### 4.2 命名空间管理
- 列表展示、新建、删除（仅 Mock）  

### 4.3 节点管理
- 列表展示、查看节点详情（标签、状态、Pod 分布）  

### 4.4 Pod 管理
- Pod 列表（名称、命名空间、状态、镜像）  
- Pod 详情（容器、日志、Events）  
- 操作：启动 / 停止 / 删除（仅 Mock）  

### 4.5 部署（Deployment）
- 列表展示、详情（YAML、Pod 状态）  
- 模拟操作：扩缩容、副本调整、镜像更新  

### 4.6 服务（Service）
- 服务列表、详情（端口映射、关联 Pod）  

### 4.7 日志与事件
- 模拟 Pod 日志输出（定时生成随机日志）  
- 事件流展示（Mock）  

---

## 5. RESTful API Mock 设计
- **集群概览**：`GET /api/cluster/overview`  
- **命名空间**：`GET /api/namespaces`，`POST /api/namespaces`，`DELETE /api/namespaces/:name`  
- **节点**：`GET /api/nodes`，`GET /api/nodes/:name`  
- **Pod**：`GET /api/pods`，`GET /api/pods/:name`，`DELETE /api/pods/:name`  
- **部署**：`GET /api/deployments`，`PUT /api/deployments/:name/scale`  
- **服务**：`GET /api/services`，`GET /api/services/:name`  

---

## 6. 界面设计要求
- **风格**：现代化，云原生主题（蓝色系）  
- **布局**：左侧导航 + 顶部工具栏 + 内容区  
- **图表**：使用 Chart.js / ECharts  

---

## 7. 非功能性需求
- Mock API 响应时间 < 300ms  
- Mock API 可替换为真实 K8s API  
- UI 易用，学习成本低  

---

## 8. 未来迭代
1. 第一阶段：UI + Mock API  
2. 第二阶段：接入真实 Kubernetes API  
3. 第三阶段：RBAC 权限控制、多集群管理  
4. 第四阶段：Helm 管理、CI/CD 可视化  

---

## 9. 附录：Mock API 示例（Go 实现）

```go
package main

import (
    "encoding/json"
    "net/http"
)

type Pod struct {
    Name      string `json:"name"`
    Namespace string `json:"namespace"`
    Status    string `json:"status"`
    Node      string `json:"node"`
    Image     string `json:"image"`
}

func getPods(w http.ResponseWriter, r *http.Request) {
    pods := []Pod{
        {"nginx-abc123", "default", "Running", "node1", "nginx:1.21"},
        {"redis-xyz789", "default", "Pending", "node2", "redis:6.2"},
        {"app-backend", "prod", "Running", "node3", "golang:1.21"},
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(pods)
}

func main() {
    http.HandleFunc("/api/pods", getPods)
    http.ListenAndServe(":8080", nil)
}
```
## 10. 开发计划（阶段与任务）

### 阶段一：UI + Mock API
- [x] 初始化 Go 项目骨架与基础 Mock 服务
- [x] 完成 `GET /api/cluster/overview` 集群概览 Mock 接口
- [x] 搭建集群概览静态页面并渲染 Mock 数据
- [x] 构建设计系统导航框架与布局基线
- [x] 扩展命名空间管理模块（Mock）
- [x] 接入节点管理数据面板（Mock）
- [x] 扩展 Pod 管理模块（Mock）
- [x] 完成部署管理模块（Mock）
- [x] 完成服务管理模块（Mock）
- [x] 提供明亮/暗黑主题切换能力
- [x] 完成日志与事件中心（Mock）
- [x] 表格体验升级（搜索 / 筛选）
- [x] 集群导入功能（kubeconfig）
- [x] 主题切换图形化提示
- [x] 事件查看按钮化（按需拉取）

### 当前迭代说明
- 目标：巩固「集群概览 + 命名空间 + 节点管理 + Deployment + Pod 管理 + Service 管理」核心体验，补齐日志与事件中心，并持续优化主题体验与表格可用性，保持 Mock API 与 UI 同步演进
- 交互：前端定时轮询（15s）刷新概览信息，支持错误提示与手动刷新；命名空间支持创建/删除反馈；节点列表支持点击查看详情；Deployment 支持模拟扩缩容；Pod 页面提供日志与事件预览；服务模块支持表格浏览、详情面板与端口/关联 Pod 呈现；日志中心提供实时日志与事件筛选；主题开关允许用户在亮色/暗色之间实时切换并记忆偏好；表格体验规划引入搜索筛选与快捷操作
- 布局：侧边导航 + 顶部工具栏 + 内容卡片区 + 节点/Deployment/Pod/Service/Logs 详情面板，统一玻璃拟态主题并适配不同亮度背景
- 技术要点：Go `net/http` + 原生前端静态页面，单体二进制即可直接运行
- 数据面：命名空间 Mock 使用内存存储，提供 `GET/POST/DELETE /api/namespaces`；节点 Mock 提供 `GET /api/nodes`、`GET /api/nodes/:name`；Deployment Mock 提供 `GET /api/deployments`、`GET /api/deployments/:name`、`PUT /api/deployments/:name/scale`；Pod Mock 提供 `GET /api/pods`、`GET /api/pods/:name`；服务模块提供 `GET /api/services`、`GET /api/services/:name`；日志中心提供 `GET /api/logs/stream`、`GET /api/events` Mock 数据，后续逐步增强表格搜索接口支持

### 服务管理模块需求与设计
- **目标**：补齐核心导航中的服务管理能力，通过 Mock 数据展示服务列表与详情，串联 Deployment 与 Pod 视图。
- **后端**：实现 `GET /api/services`、`GET /api/services/:name` 接口，Mock 数据覆盖 ClusterIP、NodePort、LoadBalancer 等类型，响应延迟控制在 200ms 以内。
- **数据要素**：服务名称、命名空间、类型、集群 IP、外部 IP、端口映射、选择器、关联 Pod 摘要、创建时间与状态徽标。
- **前端**：新增左侧导航入口「服务」，提供列表+详情联动；列表支持类型/命名空间徽章，详情呈现端口表格与关联 Pod 标签。
- **任务拆分执行进度**：
  1. ✅ 定义服务 Mock store 结构并填充种子数据；
  2. ✅ 输出服务列表与单项详情 handler，保证 REST 约定一致；
  3. ✅ 构建服务页 UI（导航、列表、详情面板）并接入 Mock 数据；
  4. ✅ 编写 store 与 handler 单元测试，保持覆盖率节奏。

### 进度速览
- ✅ 集群概览卡片与事件流已完成，Mock 数据 15s 轮询刷新
- ✅ 命名空间支持列表、创建、删除，错误信息中文提示
- ✅ 节点面板覆盖列表、详情、资源指标与条件展示
- ✅ Deployment 列表与详情可用，提供模拟扩缩容操作反馈
- ✅ Pod 列表支持日志和事件预览，涵盖调度状态
- ✅ 服务管理模块交付完成，列表/详情联动与端口、关联 Pod 展示可用
- ✅ 支持主题一键切换（亮色/暗色）并记忆偏好，亮色模式对比度显著提升
- ✅ 日志与事件中心完成，提供筛选、实时刷新与事件时间轴
- ✅ 表格体验升级交付，列表支持搜索筛选与结果提示
- ✅ 集群导入支持 kubeconfig 上传解析，保留历史记录
- ✅ 主题切换按钮图形化并提供中文提示
- ✅ 事件查看改为按需按钮触发，减少界面干扰

### 主题体验增强需求与设计
- **目标**：解决亮色环境下玻璃拟态层叠导致的灰蒙效果，提供一键明暗切换并记忆用户偏好。
- **设计要点**：
  1. ✅ 定义统一的 CSS 变量层（字体、背景、卡片、边框、文本等）；
  2. ✅ 为 `light` / `dark` 两套主题提供独立变量集合，通过 `data-theme` 控制；
  3. ✅ 顶栏加入主题开关按钮，首选取浏览器 `prefers-color-scheme`，用户选择持久化到 `localStorage`；
  4. ✅ 优化卡片透明度与文字对比，避免亮色模式下的灰蒙遮罩。
- **任务拆分**：
  1. ✅ 调整 CSS 变量与结构，明确明暗模式样式；
  2. ✅ 实现主题切换按钮与状态存储；
  3. ✅ 验证各页面在两种模式下的可读性与对比度。

### 日志与事件中心需求与设计
- **目标**：集中呈现集群运行日志与事件流，辅助运维快速定位故障并追踪Pod行为。
- **核心场景**：
  - 实时滚动展示近期 Pod 日志（Mock 数据定时刷新或手动刷新）；
  - 支持按命名空间、Pod、级别筛选日志；
  - 展示事件时间轴，包含类型、原因、消息及发生时间；
  - 提供手动刷新与自动轮询开关，保持负载可控。
- **后端**：
  - `GET /api/logs/stream?namespace=&pod=&level=` 返回最近若干条日志；
  - `GET /api/events` 返回聚合后的事件列表，按时间倒序；
  - Mock 数据内置多命名空间、多严重级别示例，响应时间 < 200ms。
- **前端交互**：
  - 新增导航「日志事件」，页面分为日志流 + 事件时间轴两列；
  - 日志区域支持级别徽章、高亮、暂停自动滚动；
  - 事件区提供类型过滤、时间戳友好展示；
  - 顶部保留主题/刷新控件，可额外新增自动刷新开关。
- **任务拆分执行进度**：
  1. ✅ 设计并实现日志/事件 Mock Store 与 Handler；
  2. ✅ 构建前端筛选控件、日志表与事件时间轴；
  3. ✅ 接入轮询与手动刷新逻辑，保证状态提示友好；
  4. ✅ 编写对应单元测试，覆盖 Mock Store 与 HTTP 层。

### 表格体验升级需求与设计
- **目标**：在核心列表（命名空间、节点、Deployment、Service、Pod）中提供快捷搜索与筛选能力，提升定位效率。
- **设计要点**：
  1. 为各数据表加入统一的搜索输入框，支持模糊匹配名称、命名空间、状态等关键字段；
  2. 复用已有缓存数据在前端本地筛选，避免额外网络开销；
  3. 搜索结果为空时提供友好提示，并保留快速清空入口；
  4. 与夜间/日间主题保持视觉一致，搜索框样式沿用输入控件体系。
- **任务拆分执行进度**：
  1. ✅ 调整前端缓存结构，确保各模块保留最新数据集；
  2. ✅ 为五大表格新增搜索栏与样式；
  3. ✅ 实现前端筛选逻辑与反馈文案；
  4. ✅ 验证交互表现并迭代用户提示信息。

### 集群导入与事件查看优化需求
- **目标**：支持通过 kubeconfig 导入集群信息，同时优化主题与事件交互体验。
- **导入设计**：
  1. 支持在集群概览页上传 kubeconfig，解析集群、上下文等核心元数据；
  2. 后端 Mock 接口提供导入摘要与历史记录；
  3. 失败时返回明确错误提示，保证文件安全性。
- **主题交互**：
  1. 主题切换按钮升级为图形化（太阳/月亮）呈现；
  2. 同步显示中文提示文案，便于理解当前模式。
- **事件查看**：
  1. 各功能页默认不展示事件列表，改用按钮按需打开；
  2. 事件数据在用户触发时加载，避免信息干扰。
- **任务拆分执行进度**：
  1. ✅ 实现 kubeconfig 解析与导入存储；
  2. ✅ 集群概览页接入导入 UI 与历史列表；
  3. ✅ 主题切换按钮图形化并更新提示；
  4. ✅ 事件查看抽屉/弹窗化，统一按钮交互。

### 快速开始
1. `go run ./cmd/dashboard` 启动本地服务，默认监听 `:8080`
2. 打开浏览器访问 `http://localhost:8080` 查看集群概览界面
3. 在左侧切换至「命名空间」模块体验创建/删除操作（纯 Mock，不持久化）
4. 在「节点管理」模块查看节点列表并点击查看详情面板
5. 在「部署与发布」模块体验 Deployment 列表、详情与模拟扩缩容
6. 切换到「Pod 管理」模块，体验日志与事件展示（纯 Mock）
7. 运行 `go test ./...` 校验 Mock 数据与接口逻辑

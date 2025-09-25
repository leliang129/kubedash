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
### 当前迭代说明
- 目标：完成「集群概览 + 命名空间 + 节点管理 + Pod 管理」核心体验，保持 Mock API 与 UI 同步演进
- 交互：前端定时轮询（15s）刷新概览信息，支持错误提示与手动刷新；命名空间支持创建/删除反馈；节点列表支持点击查看详情；Pod 页面提供日志与事件预览
- 布局：侧边导航 + 顶部工具栏 + 内容卡片区 + 节点/Pod 详情面板，统一玻璃拟态主题
- 技术要点：Go `net/http` + 原生前端静态页面，单体二进制即可直接运行
- 数据面：命名空间 Mock 使用内存存储，提供 `GET/POST/DELETE /api/namespaces`；节点 Mock 提供 `GET /api/nodes`、`GET /api/nodes/:name`；Pod Mock 提供 `GET /api/pods`、`GET /api/pods/:name`

### 快速开始
1. `go run ./cmd/dashboard` 启动本地服务，默认监听 `:8080`
2. 打开浏览器访问 `http://localhost:8080` 查看集群概览界面
3. 在左侧切换至「命名空间」模块体验创建/删除操作（纯 Mock，不持久化）
4. 在「节点管理」模块查看节点列表并点击查看详情面板
5. 切换到「Pod 管理」模块，体验日志与事件展示（纯 Mock）
6. 运行 `go test ./...` 校验 Mock 数据与接口逻辑

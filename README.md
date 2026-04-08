# saz-tool

一个面向 **Fiddler `.saz`** 的跨平台 CLI 工具，目标是把抓包导出为对 AI 和人工分析都更友好的 **normalized bundle**，同时**不破坏原始内容**。

## 当前定位

第一版先做 3 个命令边界：

- `normalize`，已实现基础版
- `show`，已实现最小可用版
- `search`，已实现最小可用版

## 设计原则

- 保留原始 SAZ，不替换原始 `raw/*` 文件
- 保留 **Fiddler session ID**
- 明确保留 **session 先后顺序**，并写入 `manifest.json`
- normalized 只做最小必要变换：
  - dechunk
  - gzip / deflate / br 解压
  - charset decode
- **JSON 不默认 pretty-print**，canonical normalized 文本保持解码后的原始文本
- 输出必须显式说明这是从 Fiddler SAZ 派生的 normalized bundle

## 输出结构

```text
example.saz.norm/
  README.md
  manifest.json
  sessions/
    000001/
      meta.json
      request.raw.txt
      response.raw.txt
      response.body.decoded.txt   # 仅当 body 可视为文本时存在
```

## Fiddler ID 与顺序

- `manifest.json.fiddlerSessionOrder` 记录所有 session id 的顺序数组
- `manifest.json.sessions[].ordinal` 表示该 session 在 bundle 内的顺序号
- `sessions/<id>/meta.json` 中同时记录：
  - `sessionId`
  - `ordinal`
  - `sourceRequestPath`
  - `sourceResponsePath`
  - `transforms`

默认第一版按 **升序 session id** 作为顺序，这通常和 Fiddler 的自然捕获顺序一致。后续若需要，也可以补充从 `raw/*_m.xml` 读取更精细时间戳顺序。

## 构建

### 本机构建

```bash
go build -o ./bin/saztool ./cmd/saztool
```

### Windows 构建

本地仍可用：

```bash
./build-windows.sh
```

但仓库现在已经提供 **GitHub Actions** 作为主构建路径：
- `.github/workflows/build-windows.yml`
- `.github/workflows/release.yml`

### Release 自动化

- push `v*` tag 时会触发 release workflow
- workflow 会构建 release assets 并上传到 GitHub Release
- 当前 release assets 覆盖：
  - windows amd64
  - windows arm64
  - linux amd64
  - linux arm64

## 使用

```bash
saztool normalize demo.saz
saztool normalize demo.saz -out demo.norm
saztool show demo.saz.norm 123
saztool search demo.saz.norm keyword
saztool search demo.saz.norm token --after-id 100 --before-id 200
saztool search demo.saz.norm token --in body
saztool search demo.saz.norm token --in request,response
saztool search demo.saz.norm token --in all
saztool search demo.saz.norm token --in request -C 2
```

更完整的参数与语义见：
- `docs/CLI.md`
- `docs/INTEGRATIONS.md`

## 说明

当前版本已落下：
- SAZ zip 读取
- raw session 聚合
- response 传输层解码
- manifest / meta / README 生成
- search scope 控制
- 上下文行搜索
- 更精确的命中展示
- session 状态标记
- timeline order
- GitHub Actions 构建与 release 自动化基础设施

还待补：
- 更多 binary body 导出策略
- 将 before/after 过滤同时支持时间顺序语义
- 直接支持 raw 目录输入
- 更丰富的 release / packaging 文档

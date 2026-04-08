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

### 交叉编译 Windows x64

```bash
GOOS=windows GOARCH=amd64 go build -o ./bin/saztool-windows-amd64.exe ./cmd/saztool
```

### 交叉编译 Windows arm64

```bash
GOOS=windows GOARCH=arm64 go build -o ./bin/saztool-windows-arm64.exe ./cmd/saztool
```

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
```

## 说明

当前版本已落下：
- SAZ zip 读取
- raw session 聚合
- response 传输层解码
- manifest / meta / README 生成

还待补：
- 更严格的 Fiddler `m.xml` 时间顺序解析
- 更多 binary body 导出策略
- 搜索结果高亮与更细的字段过滤
- 将 before/after 过滤同时支持时间顺序语义

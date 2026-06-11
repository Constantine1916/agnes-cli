# agnes-cli

**中文** | [English](https://github.com/Constantine1916/agnes-cli/blob/main/README.en.md)

面向 Agent 的 Agnes 多模态 CLI，用于通过 API Key 调用 Agnes 的图片生成和视频生成能力。

这个 CLI 的设计目标是让人类和 AI Agent 都容易调用：命令稳定、无需登录、支持环境变量/API Key 注入，成功结果只输出 URL 到 stdout，进度和诊断信息输出到 stderr。

## 安装

```bash
npm install -g @agnes-ai/agnes-cli
agnes --version
agnes doctor --offline
```

也可以从源码构建：

```bash
go build -o bin/agnes .
./bin/agnes doctor --offline
```

## 配置 API Key

Agnes CLI 不需要登录。API Key 的优先级如下：

1. `--api-key`
2. `AGNES_API_KEY`
3. `agnes key set` 保存的 key

保存 API Key：

```bash
agnes key set "$AGNES_API_KEY"
agnes key status
```

临时注入 API Key：

```bash
AGNES_API_KEY="你的_API_KEY" agnes image generate \
  --prompt "A clean product photo of a glass cube" \
  --size 1024x768
```

清除本地保存的 key：

```bash
agnes key clear
```

## 生成图片

```bash
agnes image generate \
  --prompt "A clean product photo of a glass cube" \
  --size 1024x768
```

使用图片输入：

```bash
agnes image generate \
  --prompt "Make it cinematic" \
  --image ./input.png \
  --model agnes-image-2.1-flash
```

支持的图片模型：

- `agnes-image-2.1-flash`，默认值
- `agnes-image-2.0-flash`

本地图片路径会在请求前自动转成 Data URI。

## 生成视频

```bash
agnes video generate \
  --prompt "A cat walking on the beach at sunset" \
  --num-frames 121 \
  --frame-rate 24
```

关键帧模式：

```bash
agnes video generate \
  --prompt "Smooth transition between keyframes" \
  --image ./start.png \
  --image ./end.png \
  --mode keyframes
```

查询任务状态：

```bash
agnes video status <video_id_or_task_id>
```

## Agent 友好输出约定

生成类命令成功时，stdout 只输出最终结果 URL，方便 Agent 直接解析：

```text
https://example.com/result.png
```

进度、任务 ID、诊断信息会输出到 stderr。错误会以稳定 JSON envelope 输出到 stderr，包含 `type`、`subtype`、`message` 和 `hint` 字段。

## Dry Run

使用 `--dry-run` 查看请求 payload，不会发起网络请求：

```bash
agnes --dry-run image generate \
  --prompt "A futuristic city" \
  --size 1024x768
```

## Schema

给 Agent 读取命令 schema：

```bash
agnes schema image.generate
agnes schema video.generate
```

## 常用命令

```bash
agnes key set <api-key>
agnes key status
agnes key clear

agnes image generate --prompt "..." --size 1024x768
agnes video generate --prompt "..." --num-frames 121 --frame-rate 24
agnes video status <video_id_or_task_id>
agnes doctor
```

## npm 包内容

发布到 npm 的包名是 `@agnes-ai/agnes-cli`。npm 包只包含 README、最小安装脚本和 `npm-bundles/` 下的预编译二进制压缩包；Go 源码只开源在 GitHub 仓库中。

## 发布

发版由 Git tag 触发。推送 `v0.1.0` 这样的 tag 后，GitHub Actions 会：

1. 校验版本和测试
2. 用 GoReleaser 构建 macOS、Linux、Windows 二进制
3. 上传 GitHub Release 资产和 `SHA256SUMS`
4. 将 release 资产打入 npm 包
5. 通过 npm Trusted Publishing 发布到 npm

```bash
npm version patch --no-git-tag-version
git add package.json package-lock.json internal/buildinfo/buildinfo.go
git commit -m "Release v0.1.1"
git tag v0.1.1
git push origin main --tags
```

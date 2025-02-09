# reqmd 🚀

[![Go Report Card](https://goreportcard.com/badge/github.com/voedger/reqmd)](https://goreportcard.com/report/github.com/voedger/reqmd)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Build Status](https://github.com/voedger/reqmd/actions/workflows/build.yml/badge.svg)](https://github.com/voedger/reqmd/actions)

📖 **reqmd** is a command-line tool that traces **requirements** from Markdown documents to their **coverage in source code**. It automatically generates traceability links, ensuring seamless tracking between specifications and implementation.

---

## ✨ Features

✅ Extracts **requirement references** from Markdown files  
✅ Scans **source files** for coverage tags  
✅ Generates and updates **coverage footnotes** in Markdown  
✅ Maintains a **reqmdfiles.json** file for tracking file hashes  
✅ **Fast & scalable** – uses Go concurrency to process files efficiently

---

## 🚀 Installation

Install `reqmd` via **Go**:

```sh
go install github.com/voedger/reqmd@latest
```

or build from source:

```sh
git clone https://github.com/voedger/reqmd.git
cd reqmd
go build -o reqmd .
```

---

## 🔍 Usage

### 1️⃣ **Tracing requirements**

Scan Markdown files and source repositories to generate coverage mapping:

```sh
reqmd trace <path-to-markdowns> [<path-to-cloned-repo>...]
```

#### **Arguments:**

- `<path-to-markdowns>` (**Required**) – Directory containing Markdown files  
- `<path-to-cloned-repo>` (**Optional**) – Directory containing cloned source repositories  

### 2️⃣ **Example Workflow**

#### **Markdown File (`requirements.md`)**

```markdown
- APIv2 implementation shall provide a handler for POST requests. `~Post.handler~`coverage[^~Post.handler~].
```

#### **Source Code (`handler.go`)**

```go
// [~server.api.v2/Post.handler~impl]
func handlePostRequest(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

#### **Generated Coverage Footnote**

```markdown
[^~Post.handler~]: `[~server.api.v2~impl]`[pkg/handler.go:42:impl](https://github.com/repo/pkg/handler.go#L42)
```

---

## 🛠 How It Works

1. **Scan** – Reads Markdown and source files, extracting **requirement references** and **coverage tags**  
2. **Analyze** – Validates requirement uniqueness, checks for missing coverage, and determines necessary updates  
3. **Apply** – Updates Markdown files with **footnotes** and **requirement annotations**  

**reqmd** ensures that requirements stay **linked to the actual code**, making it easier to track changes and maintain documentation integrity.

---

## Design

Ref. [design.md](design.md) for detailed design decisions and architecture.

---

## 📌 Requirements & Dependencies

🔹 **Go 1.18+**  
🔹 **Git** (for hashing file contents using `git hash-object`)  

---

## 🤝 Contributing

Contributions are welcome! 🚀 To get started:

1. **Fork** the repository  
2. **Clone** your fork locally  
3. **Create a branch** for your changes  
4. **Commit** and push  
5. **Open a Pull Request**  

```sh
git clone https://github.com/yourusername/reqmd.git
cd reqmd
git checkout -b feature-new-enhancement
```

---

## 📝 License

This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.

---

## 🌟 Acknowledgments

💡 Inspired by the need for **better traceability** between requirements and implementation in modern software projects.  

🚀 Happy tracing! 🎯

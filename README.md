# AutoChrome

#### 使用演示
![demo](docs/demo.gif)

### 通过将自然语言指令转换为无缝的浏览器交互来重新定义互联网冲浪。

思路来源[LaVague](https://github.com/lavague-ai/LaVague) 但这是一个**Golang版本的全新增强实现**

### 动机

- 旨在代表用户自动执行琐碎的任务。其中许多任务都是重复性的、耗时的，并且几乎不需要认知努力。通过自动化这些任务，用户可以腾出时间来做更有意义的事情，专注于真正重要的事情。

- 通过提供将自然语言查询转换为 Chromdp 代码的引擎，AutoChrome 旨在使用能够轻松地自动化轻松表达的 Web 工作流程并在浏览器上执行它们。

- 我们看到的关键用途之一是自动执行用户个人需要登录的任务，例如自动化支付账单、填写表格或从特定网站提取数据的过程。

- AutoChrome 基于开源项目构建，并利用本地（Ollama）或远程（OpenAI）大语言模型，以确保执行的透明度并确保其符合用户的利益。

### 特点

- **自然语言处理**：理解自然语言指令以执行浏览器交互。
- **浏览器操作**：与 Chromdp 无缝集成以实现 Web 浏览器自动化。
- **开源**：本代码以及其依赖的 AI Agent 开发框架 [AutoG](https://github.com/autogorg/autog) 100% 开源，以确保其透明度并确保其符合用户的利益。
- **隐私控制**：通过 Ollama 支持本地模型，例如Gemma-7b以便用户可以完全控制AI Agent并有隐私保障。
- **RAG 技术**：首先使用 Embedding 模型执行 RAG 来提取最相关的 HTML 片段，以上下文的形式提供给 LLM（因为直接完整的 HTML 代码大概率会超出上下文长度限制）。然后利用少样本学习和思想链来引出最相关的 Chromdp 代码来执行操作，而无需微调 LLM 。
- **解释执行（Interpreter）**：这是一个完整的 AI Interpreter 实现，对 Agent 对 AI 生成的代码进行解释执行，无缝的调用进程内的任意函数，可以大胆的想象实现任何操作（除了浏览器）！
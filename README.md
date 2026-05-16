# Controle de Estoque com Inteligência Artificial (Go & Fyne)

Este é um aplicativo desktop moderno para gerenciamento de estoque desenvolvido em **Go (Golang)**, utilizando a biblioteca gráfica **Fyne** e integrado à API oficial **Google GenAI** para geração de relatórios inteligentes.

O sistema salva os dados automaticamente de forma segura no diretório de configurações do usuário (`~/.config` ou equivalente no Windows/Mac), garantindo persistência e organização.

## 🚀 Funcionalidades

* **Gerenciamento Completo**: Cadastro e controle de produtos com preço, quantidade, categoria e status.
* **Categorias Integradas**: Suporte a Açougue, Hortifruti, Limpeza, Bebidas, Mercearia, Padaria e Laticínios.
* **Múltiplas Unidades**: Controle por Unidade (un) ou Quilo (kg).
* **Interface Gráfica (GUI)**: Interface limpa, rápida e nativa feita com Fyne.
* **Inteligência Artificial**: Integração com o Google GenAI para analisar o estoque e gerar relatórios.
* **Persistência de Dados**: Armazenamento automático em arquivo JSON em local seguro do sistema.

## 🛠️ Tecnologias Utilizadas

* [Go (Golang)](https://go.dev) - Linguagem de programação principal.
* [Fyne v2](https://fyne.io) - Toolkit de interface gráfica multiplataforma.
* [Google GenAI SDK](https://github.com) - Integração com modelos de IA do Google.

## 📋 Pré-requisitos

Antes de começar, você precisará ter instalado em sua máquina:
* Go (versão 1.16 ou superior)
* Compilador C (GCC / MinGW para o Fyne compilar os gráficos nativos)
* Uma chave de API do Google Gemini

## 🔧 Como Executar

1. Clone o repositório:
```bash
git clone https://github.com
cd NOME_DO_REPOSITORIO
```

2. Instale as dependências:
```bash
go mod tidy
```

3. Execute a aplicação:
```bash
go run main.go
```

## 📦 Armazenamento

Os dados do estoque são salvos automaticamente no formato JSON no caminho padrão do sistema operacional:
* **Linux/macOS**: `~/.config/meu-estoque/estoque.json`
* **Windows**: `%APPDATA%\meu-estoque\estoque.json`


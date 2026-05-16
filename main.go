package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/genai"
)

// --- DEFINIÇÕES DE TIPOS (MANTIDOS DO SEU CÓDIGO) ---

type Medida int

const (
	Unidade Medida = iota
	Quilo
)

func (u Medida) String() string {
	if u == Quilo {
		return "kg"
	}
	return "un"
}

type Status int

const (
	Disponivel Status = iota
	Indisponivel
)

func (s Status) String() string {
	if s == Disponivel {
		return "Disponível"
	}
	return "Indisponível"
}

type Categoria int

const (
	Acougue Categoria = iota
	Hortifruti
	Limpeza
	Bebidas
	Mercearia
	Padaria
	Laticinios
)

func (c Categoria) String() string {
	return [...]string{"Açougue", "Hortifruti", "Limpeza", "Bebidas", "Mercearia", "Padaria", "Laticínios"}[c]
}

type Produto struct {
	Name       string    `json:"nome"`
	Cat        Categoria `json:"categoria"`
	Price      float64   `json:"preco"`
	Quantidade float64   `json:"quantidade"`
	TipoMedida Medida    `json:"medida"`
	Stock      Status    `json:"status"`
	CreatedAt  time.Time `json:"data_cadastro"`
}

func (p Produto) String() string {
	dataTexto := p.CreatedAt.Format("02/01/2006")
	totalItem := p.Price * p.Quantidade
	return fmt.Sprintf("Nome: %-12s | Cat: %-10s | Preço: R$ %6.2f | Qtd: %g%-2s | Total: R$ %7.2f | Status: %-12s | Data: %s",
		p.Name, p.Cat.String(), p.Price, p.Quantidade, p.TipoMedida, totalItem, p.Stock.String(), dataTexto)
}

func (p Produto) TotalItem() float64 {
	return p.Price * p.Quantidade
}

// Define uma pasta fixa e segura no sistema do usuário para salvar o banco de dados
func obterCaminhoBanco() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "estoque.json" // Caso falhe, usa o diretório atual
	}
	appDir := filepath.Join(configDir, "meu-estoque")
	os.MkdirAll(appDir, 0755) // Cria a pasta ~/.config/meu-estoque se não existir
	return filepath.Join(appDir, "estoque.json")
}

// Define que o relatório vai para a pasta de Documentos e garante que ela exista
func obterCaminhoRelatorio() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "relatorio.txt" // Se falhar, salva na raiz do projeto
	}

	// Cria o caminho completo para a pasta Documentos
	pastaDocumentos := filepath.Join(homeDir, "Documents")

	// Força o Linux a criar a pasta "Documentos" caso ela não exista
	os.MkdirAll(pastaDocumentos, 0755)

	return filepath.Join(pastaDocumentos, "relatorio.txt")
}

func salvarJSON(produtos []Produto) error {
	dados, _ := json.MarshalIndent(produtos, "", "  ")
	return os.WriteFile(obterCaminhoBanco(), dados, 0644)
}

func carregarJSON() ([]Produto, error) {
	dados, err := os.ReadFile(obterCaminhoBanco())
	if err != nil {
		return nil, err
	}
	var produtos []Produto
	json.Unmarshal(dados, &produtos)
	return produtos, nil
}

func gerarRelatorioTxt(produtos []Produto) error {
	arquivo, err := os.Create(obterCaminhoRelatorio())

	if err != nil {
		return err
	}
	defer arquivo.Close()

	arquivo.WriteString("========== RELATÓRIO DETALHADO DE ESTOQUE ==========\n\n")
	resumoMap := make(map[Categoria]float64)
	totalGeral := 0.0

	for _, p := range produtos {
		arquivo.WriteString(p.String() + "\n")
		valorNoEstoque := p.Price * p.Quantidade
		resumoMap[p.Cat] += valorNoEstoque
		totalGeral += valorNoEstoque
	}

	arquivo.WriteString("\n========== RESUMO FINANCEIRO POR CATEGORIA ==========\n")
	for i := 0; i <= 6; i++ {
		c := Categoria(i)
		arquivo.WriteString(fmt.Sprintf("%-12s: R$ %10.2f\n", c.String(), resumoMap[c]))
	}
	arquivo.WriteString("----------------------------------------------------\n")
	arquivo.WriteString(fmt.Sprintf("VALOR TOTAL EM ESTOQUE: R$ %10.2f\n", totalGeral))
	return nil
}

// --- RENDERIZAÇÃO DA INTERFACE GRÁFICA ---

func main() {
	meuApp := app.New()
	janela := meuApp.NewWindow("Gerenciador de Estoque")
	janela.Resize(fyne.NewSize(850, 500))

	produtos, _ := carregarJSON()

	// Definição dos Cabeçalhos da Tabela
	cabecalhos := []string{"ID", "Nome", "Categoria", "Preço", "Qtd", "Total", "Status", "Data Cadastro"}

	// Nova Tabela Gráfica estruturada por Linhas e Colunas
	tabelaVisivel := widget.NewTable(
		// Define o número de linhas (total de produtos + 1 linha para o cabeçalho)
		func() (int, int) { return len(produtos) + 1, len(cabecalhos) },

		// Cria o elemento visual de texto para cada célula
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("Template")
			lbl.Alignment = fyne.TextAlignLeading
			return lbl
		},

		// Preenche os dados dentro de cada célula [Linha, Coluna]
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			// LINHA 0: Estiliza e preenche o Cabeçalho Fixo
			if id.Row == 0 {
				label.SetText(cabecalhos[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true} // Texto em Negrito
				return
			}

			// DEMAIS LINHAS: Preenche apenas com os valores brutos dos produtos
			p := produtos[id.Row-1] // -1 porque a linha 0 é o cabeçalho
			label.TextStyle = fyne.TextStyle{Bold: false}

			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", id.Row-1)) // ID do produto
			case 1:
				label.SetText(p.Name)
			case 2:
				label.SetText(p.Cat.String())
			case 3:
				label.SetText(fmt.Sprintf("R$ %.2f", p.Price))
			case 4:
				label.SetText(fmt.Sprintf("%g %s", p.Quantidade, p.TipoMedida.String()))
			case 5:
				label.SetText(fmt.Sprintf("R$ %.2f", p.TotalItem()))
			case 6:
				label.SetText(p.Stock.String())
			case 7:
				label.SetText(p.CreatedAt.Format("02/01/2006"))
			}
		},
	)

	// Ajusta dinamicamente a largura das colunas para os dados não ficarem espremidos
	tabelaVisivel.SetColumnWidth(0, 40)  // ID
	tabelaVisivel.SetColumnWidth(1, 140) // Nome
	tabelaVisivel.SetColumnWidth(2, 100) // Categoria
	tabelaVisivel.SetColumnWidth(3, 90)  // Preço
	tabelaVisivel.SetColumnWidth(4, 80)  // Quantidade
	tabelaVisivel.SetColumnWidth(5, 100) // Total
	tabelaVisivel.SetColumnWidth(6, 100) // Status
	tabelaVisivel.SetColumnWidth(7, 110) // Data

	// Atualiza a função interna que atualiza a tela
	atualizarInterface := func() {
		tabelaVisivel.Refresh()
	}

	// Formulário para Adicionar Produto
	inputNome := widget.NewEntry()
	inputPreco := widget.NewEntry()
	inputQtd := widget.NewEntry()

	selectCat := widget.NewSelect([]string{"Açougue", "Hortifruti", "Limpeza", "Bebidas", "Mercearia", "Padaria", "Laticínios"}, nil)
	selectCat.SetSelectedIndex(0)

	selectMedida := widget.NewSelect([]string{"un", "kg"}, nil)
	selectMedida.SetSelectedIndex(0)

	form := widget.NewForm(
		widget.NewFormItem("Nome do Produto", inputNome),
		widget.NewFormItem("Categoria", selectCat),
		widget.NewFormItem("Preço (R$)", inputPreco),
		widget.NewFormItem("Quantidade", inputQtd),
		widget.NewFormItem("Unidade de Medida", selectMedida),
	)

	btnAdicionar := widget.NewButton("Adicionar Produto", func() {
		preco, _ := strconv.ParseFloat(inputPreco.Text, 64)
		qtd, _ := strconv.ParseFloat(inputQtd.Text, 64)

		if inputNome.Text == "" || inputPreco.Text == "" || inputQtd.Text == "" {
			dialog.ShowError(fmt.Errorf("Por favor, preencha todos os campos!"), janela)
			return
		}

		p := Produto{
			Name:       inputNome.Text,
			Cat:        Categoria(selectCat.SelectedIndex()),
			Price:      preco,
			Quantidade: qtd,
			TipoMedida: Medida(selectMedida.SelectedIndex()),
			CreatedAt:  time.Now(),
		}

		if qtd > 0 {
			p.Stock = Disponivel
		} else {
			p.Stock = Indisponivel
		}

		produtos = append(produtos, p)
		salvarJSON(produtos)
		atualizarInterface()

		// Limpa os campos de texto após salvar
		inputNome.SetText("")
		inputPreco.SetText("")
		inputQtd.SetText("")
		dialog.ShowInformation("Sucesso", "Produto adicionado ao estoque!", janela)
	})

	// Operações de gerenciamento de itens selecionados
	btnRemover := widget.NewButton("Remover Selecionado", func() {
		// Caixa de diálogo nativa para pedir confirmação de remoção por ID
		dialog.ShowEntryDialog("Remover Produto", "Digite o número do índice do produto:", func(texto string) {
			idx, err := strconv.Atoi(texto)
			if err != nil || idx < 0 || idx >= len(produtos) {
				dialog.ShowError(fmt.Errorf("Índice inválido!"), janela)
				return
			}
			produtos = append(produtos[:idx], produtos[idx+1:]...)
			salvarJSON(produtos)
			atualizarInterface()
			dialog.ShowInformation("Sucesso", "Produto removido!", janela)
		}, janela)
	})

	btnEstoque := widget.NewButton("Movimentar Quantidade", func() {
		dialog.ShowEntryDialog("Ajustar Estoque", "Digite o Índice do produto:", func(idxTexto string) {
			idx, _ := strconv.Atoi(idxTexto)
			if idx < 0 || idx >= len(produtos) {
				dialog.ShowError(fmt.Errorf("Índice não encontrado!"), janela)
				return
			}
			dialog.ShowEntryDialog("Quantidade", "Quantidade (+ entrada, - saída):", func(qtdTexto string) {
				val, _ := strconv.ParseFloat(qtdTexto, 64)
				produtos[idx].Quantidade += val
				if produtos[idx].Quantidade <= 0 {
					produtos[idx].Quantidade = 0
					produtos[idx].Stock = Indisponivel
				} else {
					produtos[idx].Stock = Disponivel
				}
				salvarJSON(produtos)
				atualizarInterface()
			}, janela)
		}, janela)
	})

	btnPreco := widget.NewButton("Atualizar Preço", func() {
		dialog.ShowEntryDialog("Alterar Preço", "Digite o Índice do produto:", func(idxTexto string) {
			idx, _ := strconv.Atoi(idxTexto)
			if idx < 0 || idx >= len(produtos) {
				dialog.ShowError(fmt.Errorf("Índice não encontrado!"), janela)
				return
			}
			dialog.ShowEntryDialog("Novo Preço", "Digite o novo preço (R$):", func(precoTexto string) {
				nPreco, err := strconv.ParseFloat(precoTexto, 64)
				if err != nil || nPreco < 0 {
					dialog.ShowError(fmt.Errorf("Preço inválido!"), janela)
					return
				}
				produtos[idx].Price = nPreco
				salvarJSON(produtos)
				atualizarInterface()
				dialog.ShowInformation("Sucesso", "Preço atualizado com sucesso!", janela)
			}, janela)
		}, janela)
	})

	btnRelatorio := widget.NewButton("Gerar Relatório TXT", func() {
		err := gerarRelatorioTxt(produtos)
		if err != nil {
			dialog.ShowError(err, janela)
			return
		}
		dialog.ShowInformation("Sucesso", "Relatório 'relatorio.txt' exportado!", janela)
	})

	inputBusca := widget.NewEntry()
	inputBusca.SetPlaceHolder("Filtrar por nome do produto...")
	inputBusca.OnChanged = func(termo string) {
		if termo == "" {
			produtos, _ = carregarJSON()
		} else {
			todos, _ := carregarJSON()
			var filtrados []Produto
			for _, p := range todos {
				if strings.Contains(strings.ToLower(p.Name), strings.ToLower(termo)) {
					filtrados = append(filtrados, p)
				}
			}
			produtos = filtrados
		}
		atualizarInterface()
	}

	// Layout da Aplicação
	containerCadastro := container.NewVBox(widget.NewLabel("🆕 CADASTRO DE PRODUTO"), form, btnAdicionar)
	containerAcoes := container.NewVBox(widget.NewLabel("🛠️ AÇÕES"), btnRemover, btnEstoque, btnPreco, btnRelatorio)
	painelLateral := container.NewVBox(containerCadastro, widget.NewSeparator(), containerAcoes)

	painelCentral := container.NewBorder(
		container.NewVBox(widget.NewLabel("📦 ESTOQUE ATUAL"), inputBusca),
		nil, nil, nil,
		tabelaVisivel,
	)

	conteudoPrincipal := container.NewHSplit(painelLateral, painelCentral)
	conteudoPrincipal.Offset = 0.35

	janela.SetContent(conteudoPrincipal)
	janela.ShowAndRun()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  "AIzaSyDEieudOkWm_kXbnPWrH1X6f_yc9CTcf3M",
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text("Explain how AI works in a few words"),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Text())
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
)

// --- DECLARES GLOBAIS PARA SINCRONIZAÇÃO ---
var (
	produtosGlobais   []Produto
	tabelaVisivel     *widget.Table
	produtosFiltrados []Produto // Controla o que aparece na busca
	textoBusca        string
	janelaPrincipal   fyne.Window
)

// --- DEFINIÇÕES DE TIPOS ---
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

func (p Produto) TotalItem() float64 { return p.Price * p.Quantidade }

func (p Produto) String() string {
	dataTexto := p.CreatedAt.Format("02/01/2006")
	totalItem := p.Price * p.Quantidade
	return fmt.Sprintf("Nome: %-12s | Cat: %-10s | Preço: R$ %6.2f | Qtd: %g%-2s | Total: R$ %7.2f | Status: %-12s | Data: %s", p.Name, p.Cat.String(), p.Price, p.Quantidade, p.TipoMedida, totalItem, p.Stock.String(), dataTexto)
}

func main() {
	meuApp := app.New()
	janelaPrincipal = meuApp.NewWindow("Controle de Estoque Inteligente")
	janelaPrincipal.Resize(fyne.NewSize(900, 600))

	var err error
	produtosGlobais, err = carregarProdutos()
	if err != nil {
		produtosGlobais = []Produto{}
	}
	produtosFiltrados = produtosGlobais

	telaLista := criarTelaListagem()
	telaOperacoes := criarTelaOperacoes()
	telaRelatorio := criarTelaRelatorio()

	abas := container.NewAppTabs(
		container.NewTabItem("Listar Produtos", telaLista),
		container.NewTabItem("Operações Estoque", telaOperacoes),
		container.NewTabItem("Relatório TXT", telaRelatorio),
	)
	abas.SetTabLocation(container.TabLocationTop)

	janelaPrincipal.SetContent(abas)
	janelaPrincipal.ShowAndRun()
}

func atualizarInterface() {
	if textoBusca == "" {
		produtosFiltrados = produtosGlobais
	} else {
		produtosFiltrados = []Produto{}
		for _, p := range produtosGlobais {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(textoBusca)) {
				produtosFiltrados = append(produtosFiltrados, p)
			}
		}
	}
	if tabelaVisivel != nil {
		tabelaVisivel.Refresh()
	}
}

// --- CONSTRUTORES DE TELAS ---

func criarTelaListagem() fyne.CanvasObject {
	cabecalhos := []string{"ID", "Nome", "Categoria", "Preço", "Qtd", "Total", "Status", "Data Cadastro"}

	inputBusca := widget.NewEntry()
	inputBusca.SetPlaceHolder("Pesquisar produto por nome...")
	inputBusca.OnChanged = func(texto string) {
		textoBusca = texto
		atualizarInterface()
	}

	tabelaVisivel = widget.NewTable(
		func() (int, int) { return len(produtosFiltrados) + 1, len(cabecalhos) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("Template")
			lbl.Alignment = fyne.TextAlignLeading
			return lbl
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			if id.Row == 0 {
				label.SetText(cabecalhos[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			p := produtosFiltrados[id.Row-1]
			label.TextStyle = fyne.TextStyle{Bold: false}

			idReal := -1
			for idx, orig := range produtosGlobais {
				if orig.Name == p.Name && orig.CreatedAt.Equal(p.CreatedAt) {
					idReal = idx
					break
				}
			}

			switch id.Col {
			case 0:
				label.SetText(fmt.Sprintf("%d", idReal))
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

	tabelaVisivel.SetColumnWidth(0, 80)
	tabelaVisivel.SetColumnWidth(1, 140)
	tabelaVisivel.SetColumnWidth(2, 100)
	tabelaVisivel.SetColumnWidth(3, 90)
	tabelaVisivel.SetColumnWidth(4, 80)
	tabelaVisivel.SetColumnWidth(5, 100)
	tabelaVisivel.SetColumnWidth(6, 100)
	tabelaVisivel.SetColumnWidth(7, 110)

	topo := container.NewVBox(
		widget.NewLabelWithStyle("Estoque Atual", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		inputBusca,
	)

	return container.NewBorder(topo, nil, nil, nil, container.NewScroll(tabelaVisivel))
}

func criarTelaOperacoes() fyne.CanvasObject {
	inputNome := widget.NewEntry()
	inputNome.SetPlaceHolder("Nome do Produto")

	inputPreco := widget.NewEntry()
	inputPreco.SetPlaceHolder("Preço (Ex: 5.50)")

	inputQtd := widget.NewEntry()
	inputQtd.SetPlaceHolder("Quantidade")

	categoriasTexto := []string{"Açougue", "Hortifruti", "Limpeza", "Bebidas", "Mercearia", "Padaria", "Laticínios"}
	selectCat := widget.NewSelect(categoriasTexto, nil)
	selectCat.SetSelectedIndex(0)

	medidasTexto := []string{"un", "kg"}
	selectMedida := widget.NewSelect(medidasTexto, nil)
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
			dialog.ShowError(fmt.Errorf("Por favor, preencha todos os campos!"), janelaPrincipal)
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

		produtosGlobais = append(produtosGlobais, p)
		salvarJSON(produtosGlobais)
		atualizarInterface()

		inputNome.SetText("")
		inputPreco.SetText("")
		inputQtd.SetText("")
		dialog.ShowInformation("Sucesso", "Produto adicionado ao estoque!", janelaPrincipal)
	})

	btnRemover := widget.NewButton("Remover por ID", func() {
		dialog.ShowEntryDialog("Remover Produto", "Digite o número do ID Original do produto:", func(texto string) {
			idx, err := strconv.Atoi(texto)
			if err != nil || idx < 0 || idx >= len(produtosGlobais) {
				dialog.ShowError(fmt.Errorf("ID inválido! Use o ID Original listado na tabela."), janelaPrincipal)
				return
			}
			produtosGlobais = append(produtosGlobais[:idx], produtosGlobais[idx+1:]...)
			salvarJSON(produtosGlobais)
			atualizarInterface()
			dialog.ShowInformation("Sucesso", "Produto removido!", janelaPrincipal)
		}, janelaPrincipal)
	})

	btnEstoque := widget.NewButton("Movimentar Quantidade", func() {
		dialog.ShowEntryDialog("Ajustar Estoque", "Digite o ID Original do produto:", func(idxTexto string) {
			idx, _ := strconv.Atoi(idxTexto)
			if idx < 0 || idx >= len(produtosGlobais) {
				dialog.ShowError(fmt.Errorf("ID não encontrado!"), janelaPrincipal)
				return
			}
			dialog.ShowEntryDialog("Quantidade", "Quantidade (+ entrada, - saída):", func(qtdTexto string) {
				val, _ := strconv.ParseFloat(qtdTexto, 64)
				produtosGlobais[idx].Quantidade += val
				if produtosGlobais[idx].Quantidade <= 0 {
					produtosGlobais[idx].Quantidade = 0
					produtosGlobais[idx].Stock = Indisponivel
				} else {
					produtosGlobais[idx].Stock = Disponivel
				}
				salvarJSON(produtosGlobais)
				atualizarInterface()
				dialog.ShowInformation("Sucesso", "Movimentação executada!", janelaPrincipal)
			}, janelaPrincipal)
		}, janelaPrincipal)
	})

	btnPreco := widget.NewButton("Atualizar Preço", func() {
		dialog.ShowEntryDialog("Alterar Preço", "Digite o ID Original do produto:", func(idxTexto string) {
			idx, _ := strconv.Atoi(idxTexto)
			if idx < 0 || idx >= len(produtosGlobais) {
				dialog.ShowError(fmt.Errorf("ID não encontrado!"), janelaPrincipal)
				return
			}
			dialog.ShowEntryDialog("Novo Preço", "Digite o novo preço (R$):", func(precoTexto string) {
				nPreco, err := strconv.ParseFloat(precoTexto, 64)
				if err != nil || nPreco < 0 {
					dialog.ShowError(fmt.Errorf("Preço inválido!"), janelaPrincipal)
					return
				}
				produtosGlobais[idx].Price = nPreco
				salvarJSON(produtosGlobais)
				atualizarInterface()
				dialog.ShowInformation("Sucesso", "Preço atualizado com sucesso!", janelaPrincipal)
			}, janelaPrincipal)
		}, janelaPrincipal)
	})

	opcoesPeriodo := []string{"Todo o Estoque", "Diário", "Semanal", "Mensal"}
	selectPeriodo := widget.NewSelect(opcoesPeriodo, nil)
	selectPeriodo.SetSelectedIndex(0)

	formRelatorio := widget.NewForm(
		widget.NewFormItem("Período do Relatório:", selectPeriodo),
	)

	btnRelatorio := widget.NewButton("Gerar Relatório IA (TXT)", func() {
		periodoEscolhido := selectPeriodo.Selected
		produtosFiltradosParaRelatorio := filtraProdutosPorPeriodo(produtosGlobais, periodoEscolhido)

		if len(produtosFiltradosParaRelatorio) == 0 {
			dialog.ShowInformation("Aviso", fmt.Sprintf("Nenhuma movimentação ou produto encontrado para o período %s.", periodoEscolhido), janelaPrincipal)
			return
		}

		err := gerarRelatorioTxt(produtosFiltradosParaRelatorio, periodoEscolhido)
		if err != nil {
			dialog.ShowError(err, janelaPrincipal)
			return
		}
		dialog.ShowInformation("Sucesso", fmt.Sprintf("Relatório %s exportado com sucesso!", periodoEscolhido), janelaPrincipal)
	})

	// CORREÇÃO: Adicionado o retorno do layout visual que estava ausente
	layoutBotoes := container.NewVBox(btnAdicionar, btnRemover, btnEstoque, btnPreco, formRelatorio, btnRelatorio)
	return container.NewScroll(container.NewVBox(widget.NewLabel("Gerenciar Itens e Estoque"), form, layoutBotoes))
}

func criarTelaRelatorio() fyne.CanvasObject {
	textoRelatorio := widget.NewMultiLineEntry()
	textoRelatorio.SetPlaceHolder("Nenhum relatório gerado ainda...")
	textoRelatorio.Wrapping = fyne.TextWrapWord

	btnCarregar := widget.NewButton("Atualizar/Ler Relatório TXT", func() {
		caminhoTxt := obterCaminhoRelatorio()
		dados, err := ioutil.ReadFile(caminhoTxt)
		if err != nil {
			textoRelatorio.SetText("Erro ao abrir arquivo. Certifique-se de que clicou em 'Gerar Relatório IA' na aba de Operações.")
			return
		}
		textoRelatorio.SetText(string(dados))
	})

	return container.NewBorder(container.NewVBox(widget.NewLabel("Visualizador de Relatório"), btnCarregar), nil, nil, nil, textoRelatorio)
}

func carregarProdutos() ([]Produto, error) {
	produtos, err := carregarJSON()
	if err != nil {
		return nil, err
	}
	return produtos, nil
}

func obterCaminhoBanco() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "estoque.json"
	}
	appDir := filepath.Join(configDir, "meu-estoque")
	os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "estoque.json")
}

func obterCaminhoRelatorio() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "relatorio.txt"
	}
	pastaDocumentos := filepath.Join(homeDir, "Documents")
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

// CORREÇÃO: Alterado de 'produtos' para 'produtosFiltrados' dentro do loop
func gerarRelatorioTxt(produtosFiltrados []Produto, periodo string) error {
	arquivo, err := os.Create(obterCaminhoRelatorio())
	if err != nil {
		return err
	}
	defer arquivo.Close()

	arquivo.WriteString(fmt.Sprintf("========================= RELATÓRIO DETALHADO DE ESTOQUE (%s) ==========\n\n", strings.ToUpper(periodo)))
	resumoMap := make(map[Categoria]float64)
	totalGeral := 0.0

	for _, p := range produtosFiltrados {
		arquivo.WriteString(p.String() + "\n")
		valorNoEstoque := p.Price * p.Quantidade
		resumoMap[p.Cat] += valorNoEstoque
		totalGeral += valorNoEstoque
	}

	arquivo.WriteString("\n========================= RESUMO FINANCEIRO POR CATEGORIA ==========\n\n")
	for i := 0; i <= 6; i++ {
		c := Categoria(i)
		arquivo.WriteString(fmt.Sprintf("%-12s: R$ %10.2f\n", c.String(), resumoMap[c]))
	}
	arquivo.WriteString("\n----------------------------------------------------\n")
	arquivo.WriteString(fmt.Sprintf("VALOR TOTAL EM ESTOQUE: R$ %10.2f\n", totalGeral))
	return nil
}

func filtraProdutosPorPeriodo(produtos []Produto, periodo string) []Produto {
	var filtrados []Produto
	agora := time.Now()

	for _, p := range produtos {
		switch periodo {
		case "Diário":
			if p.CreatedAt.Year() == agora.Year() && p.CreatedAt.YearDay() == agora.YearDay() {
				filtrados = append(filtrados, p)
			}
		case "Semanal":
			if agora.Sub(p.CreatedAt) <= 7*24*time.Hour {
				filtrados = append(filtrados, p)
			}
		case "Mensal":
			if p.CreatedAt.Year() == agora.Year() && p.CreatedAt.Month() == agora.Month() {
				filtrados = append(filtrados, p)
			}
		default:
			filtrados = append(filtrados, p)
		}
	}
	return filtrados
}

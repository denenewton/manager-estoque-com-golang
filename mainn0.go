package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- DEFINIÇÕES DE TIPOS (ENUMS) ---

type Medida int

const (
	Unidade Medida = iota // 0
	Quilo                 // 1
)

func (u Medida) String() string {
	if u == Quilo {
		return "kg"
	}
	return "un"
}

type Status int

const (
	Disponivel   Status = iota // 0
	Indisponivel               // 1
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

// --- ESTRUTURA PRINCIPAL ---

type Produto struct {
	Name       string    `json:"nome"`
	Cat        Categoria `json:"categoria"`
	Price      float64   `json:"preco"`
	Quantidade float64   `json:"quantidade"`
	TipoMedida Medida    `json:"medida"`
	Stock      Status    `json:"status"`
	CreatedAt  time.Time `json:"data_cadastro"`
}

// String() agora inclui o CÁLCULO EM TEMPO REAL do valor total do item
func (p Produto) String() string {
	dataTexto := p.CreatedAt.Format("02/01/2006")
	totalItem := p.Price * p.Quantidade // Cálculo do total (Preço x Qtd)
	return fmt.Sprintf("Nome: %-12s | Cat: %-10s | Preço: R$ %6.2f | Qtd: %g%-2s | Total: R$ %7.2f | Status: %-12s | Data: %s",
		p.Name, p.Cat.String(), p.Price, p.Quantidade, p.TipoMedida, totalItem, p.Stock.String(), dataTexto)
}

// --- LEITURA SEGURA ---

var leitor = bufio.NewScanner(os.Stdin)

func lerTexto() string {
	leitor.Scan()
	return leitor.Text()
}

func lerNumero() float64 {
	leitor.Scan()
	valor, _ := strconv.ParseFloat(leitor.Text(), 64)
	return valor
}

// --- FUNÇÕES AUXILIARES ---

func salvarJSON(produtos []Produto) error {
	dados, _ := json.MarshalIndent(produtos, "", "  ")
	return os.WriteFile("estoque.json", dados, 0644)
}

func carregarJSON() ([]Produto, error) {
	dados, err := os.ReadFile("estoque.json")
	if err != nil {
		return nil, err
	}
	var produtos []Produto
	json.Unmarshal(dados, &produtos)
	return produtos, nil
}

func listarOpcoesCategorias() {
	fmt.Println("\nCategorias Disponíveis:")
	fmt.Println("0: Açougue | 1: Hortifruti | 2: Limpeza | 3: Bebidas | 4: Mercearia | 5: Padaria | 6: Laticínios")
}

// --- FUNÇÃO DE RELATÓRIO COM RESUMO FINANCEIRO COMPLETO ---

func gerarRelatorioTxt(produtos []Produto) error {
	arquivo, err := os.Create("relatorio.txt")
	if err != nil {
		return err
	}
	defer arquivo.Close()

	arquivo.WriteString("========== RELATÓRIO DETALHADO DE ESTOQUE ==========\n\n")

	// Mapa para o RESUMO FINANCEIRO por categoria
	resumoMap := make(map[Categoria]float64)
	totalGeral := 0.0

	for _, p := range produtos {
		arquivo.WriteString(p.String() + "\n")
		valorNoEstoque := p.Price * p.Quantidade
		resumoMap[p.Cat] += valorNoEstoque // Soma no grupo da categoria
		totalGeral += valorNoEstoque       // Soma no total geral
	}

	arquivo.WriteString("\n========== RESUMO FINANCEIRO POR CATEGORIA ==========\n")
	for i := 0; i <= 6; i++ {
		c := Categoria(i)
		arquivo.WriteString(fmt.Sprintf("%-12s: R$ %10.2f\n", c.String(), resumoMap[c]))
	}

	arquivo.WriteString("----------------------------------------------------\n")
	// TOTAL GERAL do estoque no final do arquivo
	arquivo.WriteString(fmt.Sprintf("VALOR TOTAL EM ESTOQUE: R$ %10.2f\n", totalGeral))

	return nil
}

// --- PROGRAMA PRINCIPAL ---

func main() {
	produtos, _ := carregarJSON()

	for {
		fmt.Println("\n1-Listar | 2-Add | 3-Remover | 4-Buscar | 5-Relatório | 6-Estoque | 7-Preço | 8-Filtrar Cat | 9-Sair")
		fmt.Print("Escolha uma opção: ")
		menu := int(lerNumero())

		switch menu {
		case 1: // LISTAGEM
			fmt.Println("\n--- ESTOQUE ATUAL ---")
			for _, f := range produtos {
				fmt.Println(f)
			}

		case 2: // ADICIONAR
			fmt.Print("Nome: ")
			nome := lerTexto()
			listarOpcoesCategorias()
			fmt.Print("Número da Categoria: ")
			catInt := int(lerNumero())
			if catInt < 0 || catInt > 6 {
				fmt.Println("❌ Categoria inválida!")
				continue
			}
			fmt.Print("Preço: ")
			preco := lerNumero()
			fmt.Print("Quantidade: ")
			qtd := lerNumero()
			fmt.Print("Medida (0-un, 1-kg): ")
			med := int(lerNumero())

			p := Produto{Name: nome, Cat: Categoria(catInt), Price: preco, Quantidade: qtd, TipoMedida: Medida(med), CreatedAt: time.Now()}
			if qtd > 0 {
				p.Stock = Disponivel
			} else {
				p.Stock = Indisponivel
			}

			produtos = append(produtos, p)
			salvarJSON(produtos)
			fmt.Println("✅ Produto salvo!")

		case 3: // REMOVER
			for i, f := range produtos {
				fmt.Printf("[%d] %s\n", i, f.Name)
			}
			fmt.Print("Índice: ")
			idx := int(lerNumero())
			if idx >= 0 && idx < len(produtos) {
				produtos = append(produtos[:idx], produtos[idx+1:]...)
				salvarJSON(produtos)
				fmt.Println("✅ Removido!")
			}

		case 4: // BUSCAR
			fmt.Print("Buscar por nome: ")
			termo := lerTexto()
			for _, f := range produtos {
				if strings.Contains(strings.ToLower(f.Name), strings.ToLower(termo)) {
					fmt.Println(f)
				}
			}

		case 5: // RELATÓRIO COMPLETO
			if err := gerarRelatorioTxt(produtos); err == nil {
				fmt.Println("📄 Relatório 'relatorio.txt' gerado com Resumo Financeiro!")
			}

		case 6: // MOVIMENTAR ESTOQUE
			for i, f := range produtos {
				fmt.Printf("[%d] %s (%g%s)\n", i, f.Name, f.Quantidade, f.TipoMedida)
			}
			fmt.Print("Índice: ")
			idx := int(lerNumero())
			if idx >= 0 && idx < len(produtos) {
				fmt.Print("Qtd (+ para entrada, - para saída): ")
				val := lerNumero()
				produtos[idx].Quantidade += val
				if produtos[idx].Quantidade <= 0 {
					produtos[idx].Quantidade = 0
					produtos[idx].Stock = Indisponivel
				} else {
					produtos[idx].Stock = Disponivel
				}
				salvarJSON(produtos)
				fmt.Println("✅ Estoque atualizado!")
			}

		case 7: // ATUALIZAR PREÇO
			for i, f := range produtos {
				fmt.Printf("[%d] %s (R$ %g)\n", i, f.Name, f.Price)
			}
			fmt.Print("Índice: ")
			idx := int(lerNumero())
			if idx >= 0 && idx < len(produtos) {
				fmt.Print("Novo preço: ")
				nPreco := lerNumero()
				produtos[idx].Price = nPreco
				salvarJSON(produtos)
				fmt.Println("✅ Preço atualizado!")
			}

		case 8: // FILTRAR CATEGORIA
			listarOpcoesCategorias()
			fmt.Print("Número da categoria para filtrar: ")
			escolha := Categoria(lerNumero())
			fmt.Printf("\n--- PRODUTOS EM %s ---\n", escolha.String())
			for _, f := range produtos {
				if f.Cat == escolha {
					fmt.Println(f)
				}
			}

		case 9: // SAIR
			fmt.Println("Saindo...")
			return
		}
	}
}

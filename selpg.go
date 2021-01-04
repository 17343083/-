package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/pflag"
	"io"
	"os"
	"strings"
)

type selpg_parm struct {
	start_page     int
	end_page       int
	lines_per_page int
	delimited_mode string
	print_dest     string
	file_name      string
}

func inti() {
	pflag.IntVar(&parm.start_page, "s", -1, "the start Page")
	pflag.IntVar(&parm.end_page, "e", -1, "the end Page")
	pflag.IntVar(&parm.lines_per_page, "l", 20, "the number of lines per page")
	f_flag = pflag.Bool("f", false, "")
	pflag.StringVar(&parm.print_dest, "d", "", "")
	pflag.Parse()

	parm.file_name = ""
	if pflag.NArg() == 1 {
		parm.file_name = pflag.Arg(0)
	}
}

func check(NArgs int) {
	if NArgs < 1 {
		fmt.Println("错误：没有文件路径")
		os.Exit(1)
	}
	if NArgs > 1 {
		fmt.Println("错误：除文件路径外有多余的参数")
		os.Exit(1)
	}

	if parm.start_page > parm.end_page {
		fmt.Println("错误：起始页大于终止页")
		os.Exit(1)
	}

	if parm.start_page < 1 {
		fmt.Println("错误：起始页不是正数")
		os.Exit(1)
	}

	if *f_flag && parm.lines_per_page != -1 {
		fmt.Println("错误：--l和--f冲突")
		os.Exit(1)
	} else if parm.lines_per_page != -1 {
		if parm.lines_per_page < 0 {
			fmt.Println("错误：每页的行数不是正数")
			os.Exit(1)
		} else {
			parm.delimited_mode = "l"
		}
	} else if parm.lines_per_page == -1 {
		parm.delimited_mode = "l"
		parm.lines_per_page = 20
	} else {
		parm.delimited_mode = "f"
	}
}

func operation() {
	fin := os.Stdin
	fout := os.Stdout

	var err error
	if parm.file_name != "" {
		fin, err = os.Open(parm.file_name)
		if err != nil {
			fmt.Println("错误：文件路径错误")
			os.Exit(1)
		}
		defer fin.Close()
	}

	cur_page := 1
	cur_line := 0
	if parm.delimited_mode == "f" {
		rd := bufio.NewReader(fin)
		for {
			page, ferr := rd.ReadString('\f')
			if (ferr != nil || ferr == io.EOF) && ferr == io.EOF && cur_page >= parm.start_page && cur_page <= parm.end_page {
				fmt.Fprintf(fout, "%s", page)
				break
			}
			page = strings.Replace(page, "\f", "", -1)
			cur_page++
			if cur_page >= parm.start_page && cur_page <= parm.end_page {
				fmt.Fprintf(fout, "%s", page)
			}
		}
	} else {
		line := bufio.NewScanner(fin)
		for line.Scan() {
			if cur_page >= parm.start_page && cur_page <= parm.end_page {
				fout.Write([]byte(line.Text() + "\n"))
			}
			cur_line++
			if cur_line == parm.lines_per_page {
				cur_page++
				cur_line = 0
			}
		}
	}
}

var parm selpg_parm
var f_flag *bool

func main() {
	inti()
	check(pflag.NArg())
	operation()
}

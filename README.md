本次实验的内容是使用 golang 开发 selpg

---
 - [程序设计](#catalog1)
 - [程序实现](#catalog2)
 - [测试结果](#catalog3)
 - [实验心得](#catalog4)
---

### <br /><span id="catalog1"><font face="楷体" color="#cc163a">程序设计</font></span>
输入指令的格式应该是`selpg --s start_page --e end_page [--f | --l lines_per_page] [ --d print_dest] [file_name]`
|参数|要求|意义|
|--|--|--|
--s|必需|起始页
--e|必需|结束页
--f|可选|根据分页符翻页
--l|可选|根据行数翻页，每页的行数
--d|可选|打印机地址
file_name|必需|文件路径
备注：`--f`和`--l`是互斥的，如果同时出现将会报错
根据上述情况，很容易想到利用一个结构体存储所有内容：

```go
type selpg_parm struct {
	start_page int //开始页数
	end_page int //结束页数
	lines_per_page int //每页行数
	delimited_type string // 'l' for lines-delimited, 'f' for form-feed-delimited default is 'l'
	print_dest string //打印机设备
	file_name string //输入途径，默认为键盘输入
}
```

selpg的过程分三步走：结构体初始化（输入）、检查输入、操作并输出。也就是采取以下模式：

```go
func main() {
	inti()
	check(pflag.NArg())
	operation()
}
```

接下来是具体的实现过程

### <br /><span id="catalog2"><font face="楷体" color="#cc163a">程序实现</font></span>
**一、结构体初始化（输入）**
首先要做的是将用户的输入，根据不同的参数，赋值到结构体中

```go
var parm selpg_parm
var f_flag *bool

//绑定变量
pflag.IntVar(&parm.start_page, "s", -1, "the start Page")
pflag.IntVar(&parm.end_page, "e", -1, "the end Page")
pflag.IntVar(&parm.lines_per_page, "l", -1, "the number of lines per page")
f_flag = pflag.Bool("f", false, "")
pflag.StringVar(&parm.print_dest, "d", "", "")
pflag.Parse()
```
起始页和结束页的初始值都被设置为-1，这会被用于之后判断输入是否合法。默认使用line模式，也就是按行数翻页，每页的行数初始值为20。但是如果直接赋值20，就会导致允许用户同时输入--f和--l的情况发生， 所以暂时设置为-1，后面在判断输入是否合法的时候再加以判断。一个全局变量f_flag用于判断用户是否使用了参数`--f`。最后使用`pflag.Parse()`赋值
然后输入文件名。利用pflag.NArg()得到没有前缀的参数数量。如果是1说明用户输入了文件路径，直接赋值即可

```go
parm.file_name = ""
if pflag.NArg() == 1 {
	parm.file_name = pflag.Arg(0)
}
```

**二、检查输入**
我判断了以下几种情况
|错误|条件|
|--|--|
除文件路径外有多余的参数|NArgs > 1
起始页大于终止页|parm.start_page > parm.end_page
起始页不是正数|parm.start_page < 1
每页的行数不是正数（在按行数翻页模式的前提下）|parm.lines_per_page < 0
`--l`和`--f`冲突|*f_flag && parm.lines_per_page != -1
关于最后一个错误我解释一下。当`*f_flag`为真时，说明用户输入了`--f`；当`parm.lines_per_page`不等于-1时，说明用户输入了'--l'，这就可以判断是否冲突。如果不冲突，且是`--l`，还需要判断一下每页的行数是否是初始值-1。如果是就赋值20（默认值）。这样就既避免了允许冲突的错误，也可以给每页的行数赋值20
代码：

```go
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
```

**三、操作并输出**
首先读取文件：

```go
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
```

然后根据不同的翻页模式，输出。在`f`模式下：

```go
cur_page := 1 //当前页
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
}
```

在`l`模式下：

```go
cur_line := 0 //当前行
else {
	line := bufio.NewScanner(fin)
	for line.Scan() { //逐行读入
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
```

### <br /><span id="catalog3"><font face="楷体" color="#cc163a">测试结果</font></span>
首先使用`go build selpg.go`进行部署。测试文件**test**有30行，每一行的内容都是行号
1、默认情况下，输出第2页
![1](https://img-blog.csdnimg.cn/20191224234608768.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2x1MTIwNjI1MjAwNw==,size_16,color_FFFFFF,t_70)
可以看到第2页从21行开始，到30行结束
2、每页的行数为10，输出第2页
![2](https://img-blog.csdnimg.cn/20191224234856649.png?x-oss-process=image/watermark,type_ZmFuZ3poZW5naGVpdGk,shadow_10,text_aHR0cHM6Ly9ibG9nLmNzZG4ubmV0L2x1MTIwNjI1MjAwNw==,size_16,color_FFFFFF,t_70)

接下来，尝试几种错误情况
3、除文件路径外有多余的参数
![3](https://img-blog.csdnimg.cn/20191224235153385.png)
4、起始页大于终止页
![4](https://img-blog.csdnimg.cn/20191224235542865.png)
5、起始页不是正数
![5](https://img-blog.csdnimg.cn/20191224235607224.png)
6、`--l`和`--f`冲突
![6](https://img-blog.csdnimg.cn/20191224235657200.png)
7、每页的行数不是正数
![7](https://img-blog.csdnimg.cn/20191224235710113.png)

### <br /><span id="catalog4"><font face="楷体" color="#cc163a">实验心得</font></span>
通过这次作业，我对go语言了解更深入了，学会了命令行实用程序开发，直观地感受到了pflag工具的方便

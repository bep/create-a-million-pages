package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogpeppe/go-internal/par"

	arg "github.com/alexflint/go-arg"
)

const markdownLorem1Kb = `
## Nymphe haec haec serpens totiens tale ecce

Lorem markdownum *Euhan*. *In* fuit magno tale torsit levis tenebat et murmure
primum et recepta hiemalibus fontem pecudesque omne aequantibus vicit per.

- Intellecta silet
- Tristis Baccho pietate multas latebat
- Quoque iunctis nivea seque
- Militiam clarique nobis tantoque plus periit illa
- Vulnere ubi suo dictum spectans magna lyncum
- Digitis et frondes inque et ambiguus in

Toxea Pallade, et lateri Caeneus debueram reddere indicium demersit passa
visendae est sua oculos, nive. Non oblitis, aram grave ab quem, saepe, et et
ferrum aequor. Effectum **lacrimas sive** sine deus heros, neque vel quoque
strenua fragilesque veniunt Aeolides cingentibus illa! Acutis re superet fluit;
comes solus tenet tamen natura **editus ferens** glomeravit Pygmaeae, volat.

## Supra summo flebam promptum multa

Caespite latuit indoluit comae sacrorum, cum sinuatur, nube est non [nec quoque
stetit](http://educta.com/vidissepraemia). Quae regia, hoc ultra admissa vix,
gaudias. 
`

type runner struct {
	NumPages      int    `help:"number of pages to create (defaults to a cool million)"`
	CreateBundles bool   `help:"create pages as bundles"`
	ContentSizeKb int    `help:"content size in kB (default 2)"`
	OutDir        string `help:"directory to write files to."`
}

func main() {
	r := &runner{
		NumPages:      1000000,
		ContentSizeKb: 2,
	}

	p := arg.MustParse(r)

	if r.OutDir == "" && len(r.OutDir) < 20 {
		p.Fail("must provide a sensible value for OutDir")
	}

	if r.NumPages < 100 {
		p.Fail("must provide a sensible value for NumPages")
	}

	_, err := os.Stat(r.OutDir)
	if err != nil {
		p.Fail(fmt.Sprintf("OutDir %q does not exist", r.OutDir))
	}

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

func (r *runner) getMarkdown() string {
	return strings.Repeat(markdownLorem1Kb, r.ContentSizeKb)
}

func (r *runner) Run() error {
	numWorkers := 50
	numPages := r.NumPages
	numSections := numPages / 500
	if numSections == 0 {
		numSections = 5
	}
	numPages -= numSections

	fmt.Printf("Creating %d sections with %d pages\n", numSections, numPages)

	frontMatterTemplate := `---
title: Title %d
---

`

	var sectionsWorker par.Work
	for i := 0; i < numSections; i++ {
		sectionsWorker.Add(i)
	}

	contentDir := filepath.Join(r.OutDir, "content")
	must(os.RemoveAll(contentDir))

	sectionsWorker.Do(numWorkers, func(x interface{}) {
		i := x.(int)

		dirname := filepath.Join(contentDir, fmt.Sprintf("section%d", i))
		must(os.MkdirAll(dirname, 0o777))

		filename := filepath.Join(dirname, "_index.md")
		frontmatter := fmt.Sprintf(frontMatterTemplate, i)
		must(os.WriteFile(filename, []byte(frontmatter+r.getMarkdown()), 0o666))
	})

	var pagesWorker par.Work
	for i := 0; i < numPages; i++ {
		pagesWorker.Add(i)
	}

	pagesWorker.Do(numWorkers, func(x interface{}) {
		i := x.(int)
		sectNum := rand.Intn(numSections)
		sectionDir := filepath.Join(contentDir, fmt.Sprintf("section%d", sectNum))

		var filename string
		if r.CreateBundles {
			bundleDir := filepath.Join(sectionDir, fmt.Sprintf("bundle%d", i))
			must(os.MkdirAll(bundleDir, 0o777))
			filename = filepath.Join(bundleDir, "index.md")
		} else {
			filename = filepath.Join(sectionDir, fmt.Sprintf("page%d.md", i))
		}

		frontmatter := fmt.Sprintf(frontMatterTemplate, i)

		must(os.WriteFile(filename, []byte(frontmatter+r.getMarkdown()), 0o666))
	})

	return nil
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

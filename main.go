package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reg    = regexp.MustCompile(".*\\[\\[(.*)\\]\\].*")
	regPNG = regexp.MustCompile(".*\\[\\[(.*\\.png)\\]\\].*")
)

var (
	noteName  = flag.String("n", "", "-n имя_заметки")
	vaultPath = flag.String("vault", ".", "-vault путь к vault")
)

func main() {
	flag.Parse()

	allNotesPaths := getAllNotesPaths(*noteName, nil)

	if len(allNotesPaths) == 0 {
		log.Fatal("No notes found")
	}

	tempDir, err := os.MkdirTemp(".", "")
	if err != nil {
		log.Fatal(err)
	}

	copyNotes(allNotesPaths, tempDir)
	fixImageLinks(tempDir)
	fixNewLines(tempDir)
}

func copyNotes(allNotesPaths []string, tempDir string) {
	for _, notePath := range allNotesPaths {
		newFilePath := filepath.Join(tempDir, filepath.Base(notePath))

		err := copyFileContents(notePath, newFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func fixNewLines(tempDir string) {
	dirEntries, err := os.ReadDir(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, dirEntry := range dirEntries {
		func() {
			if dirEntry.IsDir() {
				return
			}

			ext := filepath.Ext(dirEntry.Name())
			if ext != ".md" {
				return
			}

			file, err := os.OpenFile(filepath.Join(tempDir, dirEntry.Name()), os.O_RDWR, 0)
			if err != nil {
				log.Fatal(err)
			}

			defer file.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				log.Fatal(err)
			}

			fileBytes = fix(fileBytes)

			err = file.Truncate(0)
			if err != nil {
				log.Fatal(err)
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				log.Fatal(err)
			}
			_, err = file.Write(fileBytes)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

var threeBackTicks = []byte("\n```")

func fix(fileBytes []byte) []byte {
	b := make([]byte, 0)
	splitted := bytes.Split(fileBytes, []byte("```"))

	insideCodeSection := false
	for _, part := range splitted {
		if insideCodeSection {
			temp := append(threeBackTicks, part...)
			temp = append(temp, threeBackTicks...)
			b = append(b, temp...)

			insideCodeSection = false
			continue
		}

		part = bytes.Replace(part, []byte("\n"), []byte("\n\n"), -1)
		b = append(b, part...)
		insideCodeSection = true
	}

	return b
}

func fixImageLinks(tempDir string) {
	dirEntries, err := os.ReadDir(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, dirEntry := range dirEntries {
		func() {
			if dirEntry.IsDir() {
				return
			}

			ext := filepath.Ext(dirEntry.Name())
			if ext != ".md" {
				return
			}

			file, err := os.OpenFile(filepath.Join(tempDir, dirEntry.Name()), os.O_RDWR, 0)
			if err != nil {
				log.Fatal(err)
			}

			defer file.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				log.Fatal(err)
			}

			// fileBytes = bytes.Replace(fileBytes, []byte("\n"), []byte("\n\n"), -1)

			filename := strings.TrimSuffix(filepath.Base(dirEntry.Name()), ".md")

			fileBytes = append([]byte(fmt.Sprintf("# %s\n\n", filename)), fileBytes...)
			newFileBytes := regPNG.ReplaceAllFunc(fileBytes, replaceFunc)

			err = file.Truncate(0)
			if err != nil {
				log.Fatal(err)
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				log.Fatal(err)
			}
			_, err = file.Write(newFileBytes)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func replaceFunc(in []byte) []byte {
	inString := string(in)

	out := fmt.Sprintf("![_](%s)\n", inString[3:len(inString)-2])

	return []byte(out)
}

func getAllNotesPaths(linkName string, inFiles []string) (outFiles []string) {
	switch filepath.Ext(linkName) {
	case "":
		linkName = linkName + ".md"
	default:

	}

	absPath := ""

	myFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.Name() != linkName {
			return nil
		}

		absPath, err = filepath.Abs(path)
		if err != nil {
			return err
		}

		switch filepath.Ext(linkName) {
		case ".md":
			outFiles = append(outFiles, absPath)

		default:
			exist := checkExistence(inFiles, absPath)
			if exist {
				return nil
			}

			outFiles = append(outFiles, absPath)

			return nil
		}

		return nil
	}

	err := filepath.Walk(*vaultPath, myFunc)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		submatch := reg.FindAllStringSubmatch(scanner.Text(), -1)

		if len(submatch) == 0 {
			continue
		}

		links := submatch[0][1:]

		for _, link := range links {
			exist := checkExistence(inFiles, link)
			if exist {
				continue
			}

			exist = checkExistence(outFiles, link)
			if exist {
				continue
			}

			temp := make([]string, len(inFiles)+len(outFiles))
			temp = append(temp, inFiles...)
			temp = append(temp, outFiles...)

			allSubFilenames := getAllNotesPaths(link, temp)
			outFiles = append(outFiles, allSubFilenames...)

		}

	}

	return outFiles
}

func checkExistence(inFiles []string, link string) bool {
	for _, inFile := range inFiles {
		filename := filepath.Base(inFile)
		ext := filepath.Ext(inFile)
		filename = strings.TrimSuffix(filename, ext)

		if filename == link {
			return true
		}
	}

	return false
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()

	return
}

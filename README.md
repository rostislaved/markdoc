# markdoc
Prepare .md files to be converted to pdf

Obsidian isn't capable of exporting more than one note as pdf. Markdoc and pandoc can help.

## How it works:
1. Starting from root note recursively gather all notes
2. Create new temporary directory and copy all these notes in it
3. Fix all links to .png files. We need to do it because pandoc supports only special format of image links
4. Fix new lines. We need to do it because pandoc interpret one newline as no newline, so we double it

## How to use it:
1. run markdoc: 
> markdoc -n "root_note_name.md" -vault "/home/user/Obsidian Vault/"
3. Open created directory
4. Use pandoc on all *.md files. For example:
> pandoc *.md -V mainfont="Liberation Serif"  -V geometry:a4paper --pdf-engine=lualatex -o a.pdf

-V mainfont="Liberation Serif" - font with cyrillic support  
-V geometry:a4paper - set page size  
--pdf-engine=lualatex - can be installed with: apt install texlive-full  
-o a.pdf - output file (can be a.html)
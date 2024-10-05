Scout
-----
Scout is a lightweight and fast command-line tool for analyzing codebases. It helps you get a quick overview of your project's folder structure, file types, and code statistics in seconds!

![scout](https://github.com/user-attachments/assets/26c12282-2a7e-428b-af2a-d01b75838003)


🚀 Features
-----------
- Folder Structure Analysis: Generates a detailed report of the project's folder hierarchy.
- File Type & Line Count Stats: Summarizes the number of files and lines of code per file type.
- Largest File Identification: Find the biggest files in terms of size and lines.
- Respects .gitignore: Ignores files and directories listed in your .gitignore.
- File Content Summarization: Provides content summaries with truncation for large files.

🛠️ Usage
--------
```bash
scout <path/to/codebase> <path/to/outputfile>
```

- `Codebase Path`: Path to the codebase to analyze (defaults to current directory).
- `Output File Path`: Where to save the report (defaults to a file in the current directory).

📦 Installation
---------------
1. Download the binary for your platform from the Releases.
2. Extract the file:
```bash
tar -xvzf scout_<platform>.tgz
```
3. Run the `scout` binary from the command line.

🌟 Why Scout?
-------------
Scout gives you instant insights into your project's structure, helping you stay organized and focused on what matters most — coding!

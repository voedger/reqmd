# Short research/study notes

## Are emojis searchable?

- 📌
  - gitbook: yes
  - VS Code: yes
  - github: no
- 🎯
  - gitbook: yes
  - VS Code: yes
  - github: ???

## Separate `<path-to-markdowns>` and `<path-to-cloned-repo>`

By ChatGPT:

The design choice to require separate paths for `<path-to-markdowns>` and `<path-to-cloned-repo>` is intentional and serves several practical purposes:

1. **Separation of Concerns:**  
   - **Markdown Files:** The `<path-to-markdowns>` argument points to the directory where all the requirement documentation in Markdown format resides. This is where the tool scans for requirement sites, updates footnotes, and creates/updates the `reqmdfiles.json` file. Keeping these files isolated ensures that the tool works only on files that are meant to be part of the requirements traceability process.
   - **Source Files:** The `<path-to-cloned-repo>` argument (or arguments) is used exclusively for coverage analysis. It points to one or more local clones of the repository that contain the source code with CoverageTags. This separation allows the tool to validate and update coverage mappings without conflating documentation with source code.

2. **Efficiency and Accuracy:**  
   - By processing only Markdown files from the designated `<path-to-markdowns>` directory, the tool avoids unnecessarily scanning non-Markdown files that could be present in a large repository.  
   - When analyzing coverage, using a separately specified cloned repository means the tool can focus on the files that contain the actual implementations or tests (the source files) without interference from the documentation files. This targeted approach minimizes processing overhead and potential errors.

3. **Maintenance and Update Control:**  
   - The `reqmdfiles.json` file is maintained within the Markdown directory structure. This clear demarcation prevents accidental modifications or merge conflicts that might occur if Markdown and source files were processed together in one directory.
   - It also allows for independent updates—if the source repository is updated (for example, if file hashes change), only the coverage analysis part needs to re-scan the cloned repository, without impacting the Markdown documentation processing.

4. **Clarity in Traceability:**  
   - The requirements traceability process is inherently a mapping between two different kinds of artifacts: documented requirements (in Markdown) and code coverage (in source files). Keeping their file paths separate reinforces this conceptual boundary, making it easier for users and developers to understand and maintain the linkage.

In summary, while it might seem convenient to process all files from a single directory, separating the Markdown files from the source repository is a deliberate design decision to ensure that each type of file is handled appropriately. This separation improves performance, maintains clear boundaries between documentation and code, and reduces the risk of errors in the traceability process.
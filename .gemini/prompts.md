# Useful Prompts for Gemini CLI

## 1. Create a Plan
```
I need to implement [feature description]. 
Please read the current `GEMINI.md` and `README.md`. 
Then, create a detailed implementation plan in `plan.md` following the structure in `.gemini/plan_template.md`.
```

## 2. Analyze a Bug
```
I am facing a bug where [describe bug]. 
Here are the relevant files: [list files].
Please analyze the code, identify the root cause, and propose a fix in `plan.md`.
```

## 3. Generate Tests
```
I have implemented [function/class name] in [file]. 
Please generate comprehensive unit tests for it, covering edge cases and happy paths.
Follow the existing testing pattern in the codebase.
```

## 4. Update Documentation
```
I have finished implementing [feature]. 
Please review the changes and update `README.md` and `GEMINI.md` to reflect the current state of the project.
```

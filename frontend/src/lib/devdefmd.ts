export const c_initialmd=`
## Mermaid Diagrams
\`\`\`mermaid
graph TD
    A[Start] --> B{Is it working?}
    B -- Yes --> C[Great!]
    B -- No --> D[Check Logs]
\`\`\`

## Feature Verification Table

| Feature | Support | Engine | Notes |
| :--- | :---: | :--- | :--- |
| **GFM** | ✅ | Goldmark | Tables, Tasklists |
| **Theming** | ✅ | Svelte | Dark, Light, Sepia |
| **Mermaid** | ✅ | Mermaid.js | Dynamic Rendering |
| **Highlight** | ✅ | Chroma | Inline CSS |
`;
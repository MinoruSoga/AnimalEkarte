import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{a as n,i as r,n as i,o as a,r as o,s,t as c}from"./command-CdQUzcI2.js";var l,u,d,f;e((()=>{s(),l=t(),u={title:`UI/Command`,component:c,tags:[`autodocs`],parameters:{docs:{description:{component:`cmdk ラッパ — 実消費は searchable-select.tsx の検索ポップオーバー。`}}}},d={render:()=>(0,l.jsxs)(c,{style:{maxWidth:320,border:`1px solid var(--border)`,borderRadius:6},children:[(0,l.jsx)(r,{placeholder:`マスタを検索`}),(0,l.jsxs)(a,{children:[(0,l.jsx)(i,{children:`該当なし`}),(0,l.jsxs)(o,{heading:`診療`,children:[(0,l.jsx)(n,{children:`初診料`}),(0,l.jsx)(n,{children:`再診料`}),(0,l.jsx)(n,{children:`ワクチン接種`})]}),(0,l.jsxs)(o,{heading:`処置`,children:[(0,l.jsx)(n,{children:`爪切り`}),(0,l.jsx)(n,{children:`耳掃除`})]})]})]})},f=[`Default`],d.parameters={...d.parameters,docs:{...d.parameters?.docs,source:{originalSource:`{
  render: () => <Command style={{
    maxWidth: 320,
    border: "1px solid var(--border)",
    borderRadius: 6
  }}>
      <CommandInput placeholder="マスタを検索" />
      <CommandList>
        <CommandEmpty>該当なし</CommandEmpty>
        <CommandGroup heading="診療">
          <CommandItem>初診料</CommandItem>
          <CommandItem>再診料</CommandItem>
          <CommandItem>ワクチン接種</CommandItem>
        </CommandGroup>
        <CommandGroup heading="処置">
          <CommandItem>爪切り</CommandItem>
          <CommandItem>耳掃除</CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
}`,...d.parameters?.docs?.source}}}}))();export{d as Default,f as __namedExportsOrder,u as default};
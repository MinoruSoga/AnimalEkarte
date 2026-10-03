import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./button-DzIL5X86.js";import{n as i,r as a,t as o}from"./dist-CzUK0R13.js";var s,c,l=e((()=>{i(),s=t(),c=({...e})=>(0,s.jsx)(o,{theme:`light`,className:`toaster group`,closeButton:!0,position:`top-left`,offset:72,duration:4e3,style:{"--normal-bg":`var(--popover)`,"--normal-text":`var(--popover-foreground)`,"--normal-border":`var(--border)`},...e}),c.__docgenInfo={description:``,methods:[],displayName:`Toaster`}})),u,d,f,p;e((()=>{i(),l(),n(),u=t(),d={title:`UI/Toaster`,component:c,tags:[`autodocs`],parameters:{docs:{description:{component:`アプリルートの Toaster — top-left / offset 72 / closeButton 付き。`}}}},f={render:()=>(0,u.jsxs)(u.Fragment,{children:[(0,u.jsx)(c,{}),(0,u.jsxs)(`div`,{style:{display:`flex`,gap:8},children:[(0,u.jsx)(r,{variant:`outline`,onClick:()=>a(`保存しました`),children:`success toast`}),(0,u.jsx)(r,{variant:`outline`,onClick:()=>a.error(`保存に失敗しました`),children:`error toast`})]})]})},p=[`Default`],f.parameters={...f.parameters,docs:{...f.parameters?.docs,source:{originalSource:`{
  render: () => <>
      <Toaster />
      <div style={{
      display: "flex",
      gap: 8
    }}>
        <Button variant="outline" onClick={() => toast("保存しました")}>
          success toast
        </Button>
        <Button variant="outline" onClick={() => toast.error("保存に失敗しました")}>
          error toast
        </Button>
      </div>
    </>
}`,...f.parameters?.docs?.source}}}}))();export{f as Default,p as __namedExportsOrder,d as default};
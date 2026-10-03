import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./button-DzIL5X86.js";import{a as i,c as a,i as o,n as s,o as c,r as l,s as u,t as d}from"./dialog-pqHnmBGm.js";var f,p,m,h,g;e((()=>{a(),n(),f=t(),p={title:`UI/Dialog`,component:d,tags:[`autodocs`],parameters:{docs:{description:{component:`EMR-64 の focus restore 実装を含む。DialogContent は open 毎に再マウントされる。`}}}},m={render:()=>(0,f.jsxs)(d,{children:[(0,f.jsx)(u,{asChild:!0,children:(0,f.jsx)(r,{variant:`outline`,children:`新規登録`})}),(0,f.jsxs)(s,{children:[(0,f.jsxs)(i,{children:[(0,f.jsx)(c,{children:`飼主の新規登録`}),(0,f.jsx)(l,{children:`必須項目を入力して登録してください。`})]}),(0,f.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`フォーム本体はここに配置します。`}),(0,f.jsxs)(o,{children:[(0,f.jsx)(r,{variant:`outline`,children:`キャンセル`}),(0,f.jsx)(r,{children:`登録`})]})]})]})},h={render:()=>(0,f.jsx)(d,{defaultOpen:!0,children:(0,f.jsxs)(s,{children:[(0,f.jsxs)(i,{children:[(0,f.jsx)(c,{children:`削除の確認`}),(0,f.jsx)(l,{children:`この操作は取り消せません。`})]}),(0,f.jsxs)(o,{children:[(0,f.jsx)(r,{variant:`outline`,children:`キャンセル`}),(0,f.jsx)(r,{variant:`destructive`,children:`削除`})]})]})})},g=[`Default`,`Open`],m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  render: () => <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline">新規登録</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>飼主の新規登録</DialogTitle>
          <DialogDescription>必須項目を入力して登録してください。</DialogDescription>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">フォーム本体はここに配置します。</p>
        <DialogFooter>
          <Button variant="outline">キャンセル</Button>
          <Button>登録</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <Dialog defaultOpen>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>削除の確認</DialogTitle>
          <DialogDescription>この操作は取り消せません。</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline">キャンセル</Button>
          <Button variant="destructive">削除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
}`,...h.parameters?.docs?.source},description:{story:`初期表示で開いた状態（a11y addon がオーバーレイを検査できるように）。`,...h.parameters?.docs?.description}}}}))();export{m as Default,h as Open,g as __namedExportsOrder,p as default};
import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./button-DzIL5X86.js";import{a as i,c as a,i as o,l as s,n as c,o as l,r as u,s as d,t as f}from"./alert-dialog-DTU6X50j.js";var p,m,h,g;e((()=>{s(),n(),p=t(),m={title:`UI/AlertDialog`,component:f,tags:[`autodocs`],parameters:{docs:{description:{component:`破壊的操作の確認用。Trigger は export されていないため defaultOpen で開く。`}}}},h={render:()=>(0,p.jsx)(f,{defaultOpen:!0,children:(0,p.jsxs)(o,{children:[(0,p.jsxs)(d,{children:[(0,p.jsx)(a,{children:`カルテを削除しますか？`}),(0,p.jsx)(i,{children:`この操作は取り消せません。削除対象のカルテは完全に削除されます。`})]}),(0,p.jsxs)(l,{children:[(0,p.jsx)(u,{asChild:!0,children:(0,p.jsx)(r,{variant:`outline`,children:`キャンセル`})}),(0,p.jsx)(c,{asChild:!0,children:(0,p.jsx)(r,{variant:`destructive`,children:`削除する`})})]})]})})},g=[`Open`],h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <AlertDialog defaultOpen>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>カルテを削除しますか？</AlertDialogTitle>
          <AlertDialogDescription>
            この操作は取り消せません。削除対象のカルテは完全に削除されます。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel asChild>
            <Button variant="outline">キャンセル</Button>
          </AlertDialogCancel>
          <AlertDialogAction asChild>
            <Button variant="destructive">削除する</Button>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
}`,...h.parameters?.docs?.source}}}}))();export{h as Open,g as __namedExportsOrder,m as default};
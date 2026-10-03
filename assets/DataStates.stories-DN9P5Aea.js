import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{A as n,t as r}from"./lucide-react-yS7Xowms.js";import{n as i,t as a}from"./button-DzIL5X86.js";import{i as o,n as s,r as c,t as l}from"./DataStates-DJyH2Xjz.js";var u,d,f,p,m,h,g,_;e((()=>{r(),i(),o(),u=t(),d={title:`Shared/DataStates`,component:l,tags:[`autodocs`]},f={render:()=>(0,u.jsx)(c,{})},p={name:`Error`,render:()=>(0,u.jsx)(s,{})},m={name:`Error（カスタムメッセージ）`,render:()=>(0,u.jsx)(s,{message:`診察記録の取得に失敗しました`})},h={args:{message:`記録がありません`,description:`条件を変更して再検索してください。`,icon:(0,u.jsx)(n,{className:`size-8`})}},g={args:{message:`カルテがありません`,children:(0,u.jsx)(a,{size:`sm`,children:`記録を作成`})}},_=[`Loading`,`Error_`,`ErrorCustomMessage`,`Empty`,`EmptyWithAction`],f.parameters={...f.parameters,docs:{...f.parameters?.docs,source:{originalSource:`{
  render: () => <LoadingFallback />
}`,...f.parameters?.docs?.source}}},p.parameters={...p.parameters,docs:{...p.parameters?.docs,source:{originalSource:`{
  name: "Error",
  render: () => <ErrorFallback />
}`,...p.parameters?.docs?.source}}},m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  name: "Error（カスタムメッセージ）",
  render: () => <ErrorFallback message="診察記録の取得に失敗しました" />
}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  args: {
    message: "記録がありません",
    description: "条件を変更して再検索してください。",
    icon: <Search className="size-8" />
  }
}`,...h.parameters?.docs?.source}}},g.parameters={...g.parameters,docs:{...g.parameters?.docs,source:{originalSource:`{
  args: {
    message: "カルテがありません",
    children: <Button size="sm">記録を作成</Button>
  }
}`,...g.parameters?.docs?.source}}}}))();export{h as Empty,g as EmptyWithAction,m as ErrorCustomMessage,p as Error_,f as Loading,_ as __namedExportsOrder,d as default};
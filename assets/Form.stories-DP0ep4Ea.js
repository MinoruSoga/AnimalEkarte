import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{t}from"./jsx-runtime-DeHZSEgm.js";import{d as n,n as r}from"./design-tokens-B-kSoC24.js";import{n as i,t as a}from"./button-CKac4K-h.js";import{n as o,t as s}from"./SubmitButton-uZi3sEr2.js";import{n as c,t as l}from"./FormHeader-B9qjPxvH.js";function u({className:e,colorVariant:t=`primary`,...n}){let i=t===`brand`?`${r.bgBrandIdentity} ${r.textOnBrandIdentity} ${r.hoverBgBrandIdentity} ${r.hoverTextOnBrandIdentity} ${r.activeBgBrandIdentity} ${r.activeTextOnBrandIdentity} h-11 text-xl font-bold shadow-none rounded-full border-transparent`:`${r.bgActionPrimarySolid} ${r.textOnActionPrimary} ${r.hoverBgActionPrimarySolid} ${r.hoverTextOnActionPrimary} ${r.activeBgActionPrimarySolid} ${r.activeTextOnActionPrimary} h-11 text-base shadow-none rounded-full border-transparent`;return(0,d.jsx)(a,{className:`${i} ${e||``}`,...n})}var d;function f(){return(f=e((()=>{i(),n(),d=t(),u.__docgenInfo={description:``,methods:[],displayName:`PrimaryButton`,props:{colorVariant:{required:!1,tsType:{name:`union`,raw:`"default" | "primary" | "brand"`,elements:[{name:`literal`,value:`"default"`},{name:`literal`,value:`"primary"`},{name:`literal`,value:`"brand"`}]},description:`"primary"（既定） — 汎用の主操作（brand と同じ primary teal + pill）。
"brand" — 認証など製品識別面の Brand CTA（teal + pill）。
"default" — "primary" の後方互換 alias。
SubmitButton と対称のプロパティ名・実装（単一 className 文字列を選択し連結しない）。`,defaultValue:{value:`"primary"`,computed:!1}}},composes:[`ButtonProps`]}})))()}var p,m,h,g,_,v;function y(){return(y=e((()=>{c(),o(),f(),p=t(),m={title:`Shared/Form`,component:l,tags:[`autodocs`],args:{title:`診療内容の登録`}},h={},g={args:{title:`ワクチン接種記録`,description:`接種日とワクチン種別を入力します`,onBack:()=>{},action:(0,p.jsx)(u,{children:`保存する`})}},_={render:()=>(0,p.jsxs)(`div`,{className:`flex flex-col items-start gap-4`,children:[(0,p.jsx)(u,{children:`PrimaryButton（primary）`}),(0,p.jsx)(u,{colorVariant:`brand`,children:`PrimaryButton（brand）`}),(0,p.jsx)(u,{colorVariant:`default`,children:`PrimaryButton（default alias）`}),(0,p.jsx)(u,{disabled:!0,children:`disabled`}),(0,p.jsx)(s,{children:`SubmitButton（form 連動）`}),(0,p.jsx)(s,{colorVariant:`destructive`,children:`SubmitButton（destructive）`})]})},v=[`Header`,`HeaderFull`,`SubmitButtons`],h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{}`,...h.parameters?.docs?.source}}},g.parameters={...g.parameters,docs:{...g.parameters?.docs,source:{originalSource:`{
  args: {
    title: "ワクチン接種記録",
    description: "接種日とワクチン種別を入力します",
    onBack: () => {},
    action: <PrimaryButton>保存する</PrimaryButton>
  }
}`,...g.parameters?.docs?.source}}},_.parameters={..._.parameters,docs:{..._.parameters?.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col items-start gap-4">
      <PrimaryButton>PrimaryButton（primary）</PrimaryButton>
      <PrimaryButton colorVariant="brand">PrimaryButton（brand）</PrimaryButton>
      <PrimaryButton colorVariant="default">PrimaryButton（default alias）</PrimaryButton>
      <PrimaryButton disabled>disabled</PrimaryButton>
      <SubmitButton>SubmitButton（form 連動）</SubmitButton>
      <SubmitButton colorVariant="destructive">SubmitButton（destructive）</SubmitButton>
    </div>
}`,..._.parameters?.docs?.source}}}})))()}y();export{h as Header,g as HeaderFull,_ as SubmitButtons,v as __namedExportsOrder,m as default};
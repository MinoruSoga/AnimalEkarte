import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{K as n,t as r}from"./lucide-react-yS7Xowms.js";import{d as i,n as a}from"./design-tokens-CoIxcO1x.js";import{n as o,t as s}from"./button-DzIL5X86.js";import{n as c,t as l}from"./SubmitButton-BFCMun5G.js";import{n as u,t as d}from"./FormHeader-DOuSrtiN.js";function f({className:e,colorVariant:t=`primary`,...n}){return(0,p.jsx)(s,{className:`${t===`brand`?`${a.bgBrandIdentity} ${a.textOnBrandIdentity} ${a.hoverBgBrandIdentity} ${a.hoverTextOnBrandIdentity} ${a.activeBgBrandIdentity} ${a.activeTextOnBrandIdentity} h-11 text-xl font-bold shadow-none rounded-full border-transparent`:`${a.bgActionPrimarySolid} ${a.textOnActionPrimary} ${a.hoverBgActionPrimarySolid} ${a.hoverTextOnActionPrimary} ${a.activeBgActionPrimarySolid} ${a.activeTextOnActionPrimary} h-11 text-base shadow-none rounded-full border-transparent`} ${e||``}`,...n})}var p,m=e((()=>{o(),i(),p=t(),f.__docgenInfo={description:``,methods:[],displayName:`PrimaryButton`,props:{colorVariant:{required:!1,tsType:{name:`union`,raw:`"default" | "primary" | "brand"`,elements:[{name:`literal`,value:`"default"`},{name:`literal`,value:`"primary"`},{name:`literal`,value:`"brand"`}]},description:`"primary"（既定） — 汎用の主操作（brand と同じ primary teal + pill）。
"brand" — 認証など製品識別面の Brand CTA（teal + pill）。
"default" — "primary" の後方互換 alias。
SubmitButton と対称のプロパティ名・実装（単一 className 文字列を選択し連結しない）。`,defaultValue:{value:`"primary"`,computed:!1}}},composes:[`ButtonProps`]}})),h,g,_,v,y,b;e((()=>{r(),u(),c(),m(),h=t(),g={title:`Shared/Form`,component:d,tags:[`autodocs`],args:{title:`診療内容の登録`}},_={},v={args:{title:`ワクチン接種記録`,description:`接種日とワクチン種別を入力します`,icon:(0,h.jsx)(n,{className:`size-5`}),onBack:()=>{},action:(0,h.jsx)(f,{children:`保存する`})}},y={render:()=>(0,h.jsxs)(`div`,{className:`flex flex-col items-start gap-4`,children:[(0,h.jsx)(f,{children:`PrimaryButton（primary）`}),(0,h.jsx)(f,{colorVariant:`brand`,children:`PrimaryButton（brand）`}),(0,h.jsx)(f,{colorVariant:`default`,children:`PrimaryButton（default alias）`}),(0,h.jsx)(f,{disabled:!0,children:`disabled`}),(0,h.jsx)(l,{children:`SubmitButton（form 連動）`}),(0,h.jsx)(l,{colorVariant:`destructive`,children:`SubmitButton（destructive）`})]})},b=[`Header`,`HeaderFull`,`SubmitButtons`],_.parameters={..._.parameters,docs:{..._.parameters?.docs,source:{originalSource:`{}`,..._.parameters?.docs?.source}}},v.parameters={...v.parameters,docs:{...v.parameters?.docs,source:{originalSource:`{
  args: {
    title: "ワクチン接種記録",
    description: "接種日とワクチン種別を入力します",
    icon: <PawPrint className="size-5" />,
    onBack: () => {},
    action: <PrimaryButton>保存する</PrimaryButton>
  }
}`,...v.parameters?.docs?.source}}},y.parameters={...y.parameters,docs:{...y.parameters?.docs,source:{originalSource:`{
  render: () => <div className="flex flex-col items-start gap-4">
      <PrimaryButton>PrimaryButton（primary）</PrimaryButton>
      <PrimaryButton colorVariant="brand">PrimaryButton（brand）</PrimaryButton>
      <PrimaryButton colorVariant="default">PrimaryButton（default alias）</PrimaryButton>
      <PrimaryButton disabled>disabled</PrimaryButton>
      <SubmitButton>SubmitButton（form 連動）</SubmitButton>
      <SubmitButton colorVariant="destructive">SubmitButton（destructive）</SubmitButton>
    </div>
}`,...y.parameters?.docs?.source}}}}))();export{_ as Header,v as HeaderFull,y as SubmitButtons,b as __namedExportsOrder,g as default};
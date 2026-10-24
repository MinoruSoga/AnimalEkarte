import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{t}from"./jsx-runtime-DeHZSEgm.js";import{d as n,t as r}from"./design-tokens-B-kSoC24.js";import{n as i,t as a}from"./badge-Dw9ZSJew.js";function o({children:e,className:t=``,colorClass:n=``}){return(0,s.jsx)(a,{variant:`outline`,className:`text-2xs px-2 h-7 font-semibold border ${n} ${t}`,children:e})}var s;function c(){return(c=e((()=>{i(),s=t(),o.__docgenInfo={description:``,methods:[],displayName:`StatusBadge`,props:{children:{required:!0,tsType:{name:`ReactReactNode`,raw:`React.ReactNode`},description:``},className:{required:!1,tsType:{name:`string`},description:``,defaultValue:{value:`""`,computed:!1}},colorClass:{required:!1,tsType:{name:`string`},description:``,defaultValue:{value:`""`,computed:!1}}}}})))()}var l,u,d,f,p,m,h,g;function _(){return(_=e((()=>{n(),c(),l=t(),u={title:`Shared/StatusBadge`,component:o,tags:[`autodocs`],args:{children:`受付中`,colorClass:r.green}},d={},f={args:{children:`依頼中`,colorClass:r.yellow}},p={args:{children:`期限切れ`,colorClass:r.red}},m={args:{children:`終了`,colorClass:r.gray}},h={render:()=>(0,l.jsxs)(`div`,{className:`flex flex-wrap items-center gap-2`,children:[(0,l.jsx)(o,{colorClass:r.green,children:`受付中`}),(0,l.jsx)(o,{colorClass:r.yellow,children:`依頼中`}),(0,l.jsx)(o,{colorClass:r.orange,children:`進行中`}),(0,l.jsx)(o,{colorClass:r.red,children:`期限切れ`}),(0,l.jsx)(o,{colorClass:r.muted,children:`未対応`}),(0,l.jsx)(o,{colorClass:r.gray,children:`終了`})]})},g=[`Green`,`Yellow`,`Red`,`Gray`,`AllVariants`],d.parameters={...d.parameters,docs:{...d.parameters?.docs,source:{originalSource:`{}`,...d.parameters?.docs?.source}}},f.parameters={...f.parameters,docs:{...f.parameters?.docs,source:{originalSource:`{
  args: {
    children: "依頼中",
    colorClass: BADGE.yellow
  }
}`,...f.parameters?.docs?.source}}},p.parameters={...p.parameters,docs:{...p.parameters?.docs,source:{originalSource:`{
  args: {
    children: "期限切れ",
    colorClass: BADGE.red
  }
}`,...p.parameters?.docs?.source}}},m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  args: {
    children: "終了",
    colorClass: BADGE.gray
  }
}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <div className="flex flex-wrap items-center gap-2">
      <StatusBadge colorClass={BADGE.green}>受付中</StatusBadge>
      <StatusBadge colorClass={BADGE.yellow}>依頼中</StatusBadge>
      <StatusBadge colorClass={BADGE.orange}>進行中</StatusBadge>
      <StatusBadge colorClass={BADGE.red}>期限切れ</StatusBadge>
      <StatusBadge colorClass={BADGE.muted}>未対応</StatusBadge>
      <StatusBadge colorClass={BADGE.gray}>終了</StatusBadge>
    </div>
}`,...h.parameters?.docs?.source}}}})))()}_();export{h as AllVariants,m as Gray,d as Green,p as Red,f as Yellow,g as __namedExportsOrder,u as default};
import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{f as t}from"./iframe-C6Sk8rp0.js";import{t as n}from"./jsx-runtime-DeHZSEgm.js";import{d as r,s as i}from"./design-tokens-B-kSoC24.js";import{a,i as o,n as s,o as c,r as l,t as u}from"./select-DpqwdMWI.js";var d,f,p,m;function h(){return(h=e((()=>{d=t(),c(),r(),f=n(),p=(0,f.jsxs)(f.Fragment,{children:[(0,f.jsx)(l,{value:`excluded`,children:`外税`}),(0,f.jsx)(l,{value:`included`,children:`内税`}),(0,f.jsx)(l,{value:`exempt`,children:`非課税`})]}),m=(0,d.memo)(function({value:e,onChange:t,disabled:n,ariaLabel:r,className:c}){return(0,f.jsxs)(u,{value:e,onValueChange:e=>t(e),disabled:n,children:[(0,f.jsx)(o,{"aria-label":r,className:`${i.selectCompact} ${c??``}`.trimEnd(),children:(0,f.jsx)(a,{placeholder:`選択`})}),(0,f.jsx)(s,{children:p})]})}),m.__docgenInfo={description:``,methods:[],displayName:`TaxTypeSelector`,props:{value:{required:!0,tsType:{name:`TaxType`},description:``},onChange:{required:!0,tsType:{name:`signature`,type:`function`,raw:`(value: TaxType) => void`,signature:{arguments:[{type:{name:`TaxType`},name:`value`}],return:{name:`void`}}},description:``},disabled:{required:!1,tsType:{name:`boolean`},description:``},ariaLabel:{required:!1,tsType:{name:`string`},description:``},className:{required:!1,tsType:{name:`string`},description:``}}}})))()}var g,_,v,y,b,x;function S(){return(S=e((()=>{h(),g={title:`Shared/TaxTypeSelector`,component:m,tags:[`autodocs`],args:{value:`included`,onChange:()=>{}}},_={},v={args:{value:`excluded`}},y={args:{value:`exempt`}},b={args:{disabled:!0}},x=[`Included`,`Excluded`,`Exempt`,`Disabled`],_.parameters={..._.parameters,docs:{..._.parameters?.docs,source:{originalSource:`{}`,..._.parameters?.docs?.source}}},v.parameters={...v.parameters,docs:{...v.parameters?.docs,source:{originalSource:`{
  args: {
    value: "excluded"
  }
}`,...v.parameters?.docs?.source}}},y.parameters={...y.parameters,docs:{...y.parameters?.docs,source:{originalSource:`{
  args: {
    value: "exempt"
  }
}`,...y.parameters?.docs?.source}}},b.parameters={...b.parameters,docs:{...b.parameters?.docs,source:{originalSource:`{
  args: {
    disabled: true
  }
}`,...b.parameters?.docs?.source}}}})))()}S();export{b as Disabled,v as Excluded,y as Exempt,_ as Included,x as __namedExportsOrder,g as default};
import{i as e,s as t}from"./preload-helper-xPQekRTU.js";import{O as n}from"./iframe-Bt5VGDn5.js";import{t as r}from"./jsx-runtime-CaZkqeYb.js";import{d as i,s as a}from"./design-tokens-CoIxcO1x.js";import{a as o,i as s,n as c,o as l,r as u,t as d}from"./select-QiJ9CgwX.js";var f,p,m,h,g=e((()=>{f=t(n(),1),l(),i(),p=r(),m=(0,p.jsxs)(p.Fragment,{children:[(0,p.jsx)(u,{value:`excluded`,children:`外税`}),(0,p.jsx)(u,{value:`included`,children:`内税`}),(0,p.jsx)(u,{value:`exempt`,children:`非課税`})]}),h=(0,f.memo)(function({value:e,onChange:t,disabled:n,ariaLabel:r,className:i}){return(0,p.jsxs)(d,{value:e,onValueChange:e=>t(e),disabled:n,children:[(0,p.jsx)(s,{"aria-label":r,className:`${a.selectCompact} ${i??``}`.trimEnd(),children:(0,p.jsx)(o,{placeholder:`選択`})}),(0,p.jsx)(c,{children:m})]})}),h.__docgenInfo={description:``,methods:[],displayName:`TaxTypeSelector`,props:{value:{required:!0,tsType:{name:`TaxType`},description:``},onChange:{required:!0,tsType:{name:`signature`,type:`function`,raw:`(value: TaxType) => void`,signature:{arguments:[{type:{name:`TaxType`},name:`value`}],return:{name:`void`}}},description:``},disabled:{required:!1,tsType:{name:`boolean`},description:``},ariaLabel:{required:!1,tsType:{name:`string`},description:``},className:{required:!1,tsType:{name:`string`},description:``}}}})),_,v,y,b,x,S;e((()=>{g(),_={title:`Shared/TaxTypeSelector`,component:h,tags:[`autodocs`],args:{value:`included`,onChange:()=>{}}},v={},y={args:{value:`excluded`}},b={args:{value:`exempt`}},x={args:{disabled:!0}},S=[`Included`,`Excluded`,`Exempt`,`Disabled`],v.parameters={...v.parameters,docs:{...v.parameters?.docs,source:{originalSource:`{}`,...v.parameters?.docs?.source}}},y.parameters={...y.parameters,docs:{...y.parameters?.docs,source:{originalSource:`{
  args: {
    value: "excluded"
  }
}`,...y.parameters?.docs?.source}}},b.parameters={...b.parameters,docs:{...b.parameters?.docs,source:{originalSource:`{
  args: {
    value: "exempt"
  }
}`,...b.parameters?.docs?.source}}},x.parameters={...x.parameters,docs:{...x.parameters?.docs,source:{originalSource:`{
  args: {
    disabled: true
  }
}`,...x.parameters?.docs?.source}}}}))();export{x as Disabled,y as Excluded,b as Exempt,v as Included,S as __namedExportsOrder,_ as default};
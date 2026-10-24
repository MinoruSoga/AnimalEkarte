import{a as e,n as t}from"./rolldown-runtime-DkW27tQK.js";import{f as n}from"./iframe-C6Sk8rp0.js";import{t as r}from"./jsx-runtime-DeHZSEgm.js";import{i,n as a,r as o,t as s}from"./dist-CbFlsgBU.js";import{n as c,t as l}from"./dist-S0mBbiUy.js";import{r as u,t as d}from"./dist-oJCNPtvx.js";import{n as f,t as p}from"./dist-DIeBUw0P.js";import{n as m,t as h}from"./dist-BUw9JkLV.js";import{n as g,t as _}from"./dist-BW9ECE_1.js";import{n as v,t as y}from"./utils-DT5hf25b.js";import{d as b,n as x}from"./design-tokens-B-kSoC24.js";import{n as S,t as C}from"./check-C__z63eu.js";import{n as w,t as T}from"./label-Ct3Dkfk1.js";function E(e){let{__scopeCheckbox:t,checked:n,children:r,defaultChecked:i,disabled:a,form:o,name:s,onCheckedChange:c,required:l,value:u=`on`,internal_do_not_use_render:d}=e,[f,p]=g({prop:n,defaultProp:i??!1,onChange:c,caller:P}),[m,h]=A.useState(null),[_,v]=A.useState(null),[y,b]=A.useState(null),x=!m||!!o||!!m.closest(`form`),S={checked:f,disabled:a,setChecked:p,control:m,setControl:h,name:s,form:o,value:u,userClick:y,onUserClick:b,required:l,defaultChecked:!O(i)&&i,isFormControl:x,bubbleInput:_,setBubbleInput:v};return(0,j.jsx)(I,{scope:t,...S,children:D(d)?d(S):r})}function D(e){return typeof e==`function`}function O(e){return e===`indeterminate`}function k(e){return O(e)?`indeterminate`:e?`checked`:`unchecked`}var A,j,M,N,P,F,ee,I,L,R,z,B,V,H,U,W;function G(){return(G=t((()=>{A=e(n(),1),l(),a(),i(),_(),p(),m(),u(),j=r(),M=Object.defineProperty,N=(e,t)=>M(e,`name`,{value:t,configurable:!0}),P=`Checkbox`,[F,ee]=s(P),[I,L]=F(P),N(E,`CheckboxProvider`),R=`CheckboxTrigger`,z=A.forwardRef(N(function({__scopeCheckbox:e,onKeyDown:t,onClick:n,...r},i){let{control:a,value:s,disabled:l,checked:u,required:f,setControl:p,setChecked:m,onUserClick:h,isFormControl:g,bubbleInput:_}=L(R,e),v=c(i,p),y=A.useRef(u);return A.useEffect(()=>{let e=a?.form;if(e){let t=N(()=>m(y.current),`reset`);return e.addEventListener(`reset`,t),()=>e.removeEventListener(`reset`,t)}},[a,m]),(0,j.jsx)(d.button,{type:`button`,role:`checkbox`,"aria-checked":O(u)?`mixed`:u,"aria-required":f,"data-state":k(u),"data-disabled":l?``:void 0,disabled:l,value:s,...r,ref:v,onKeyDown:o(t,e=>{e.key===`Enter`&&e.preventDefault()}),onClick:o(n,e=>{let t=e.isPropagationStopped();h({requestedChecked:O(u)?!0:!u,hasConsumerStoppedPropagation:t}),m(e=>O(e)?!0:!e),_&&g&&(t||e.stopPropagation())})})},`CheckboxTrigger`)),B=A.forwardRef(N(function(e,t){let{__scopeCheckbox:n,name:r,checked:i,defaultChecked:a,required:o,disabled:s,value:c,onCheckedChange:l,form:u,...d}=e;return(0,j.jsx)(E,{__scopeCheckbox:n,checked:i,defaultChecked:a,disabled:s,required:o,onCheckedChange:l,name:r,form:u,value:c,internal_do_not_use_render:({isFormControl:e})=>(0,j.jsxs)(j.Fragment,{children:[(0,j.jsx)(z,{...d,ref:t,__scopeCheckbox:n}),e&&(0,j.jsx)(W,{__scopeCheckbox:n})]})})},`Checkbox`)),V=`CheckboxIndicator`,H=A.forwardRef(N(function(e,t){let{__scopeCheckbox:n,forceMount:r,...i}=e,a=L(V,n);return(0,j.jsx)(h,{present:r||O(a.checked)||a.checked===!0,children:(0,j.jsx)(d.span,{"data-state":k(a.checked),"data-disabled":a.disabled?``:void 0,...i,ref:t,style:{pointerEvents:`none`,...e.style}})})},`CheckboxIndicator`)),U=`CheckboxBubbleInput`,W=A.forwardRef(N(function({__scopeCheckbox:e,onClick:t,...n},r){let{control:i,userClick:a,checked:s,defaultChecked:l,required:u,disabled:p,name:m,value:h,form:g,bubbleInput:_,setBubbleInput:v}=L(U,e),y=c(r,v),b=f(i),x=A.useRef(!1),S=A.useRef(s),C=A.useRef(a),w=A.useRef(null);A.useEffect(()=>{let e=_;if(!e)return;let t=window.HTMLInputElement.prototype,n=Object.getOwnPropertyDescriptor(t,`checked`).set;if(!n)return;let r=S.current!==s;S.current=s;let i=a!==null&&a!==C.current;if(C.current=a,i)w.current=r?null:a,r&&(e.indeterminate=O(s),n.call(e,!O(s)&&s)),e.dispatchEvent(new Event(`click`,{bubbles:!a.hasConsumerStoppedPropagation}));else if(r){let t=w.current;w.current=null;let r=t?.requestedChecked===s;x.current=!0,e.indeterminate=O(s),n.call(e,!O(s)&&s),e.dispatchEvent(new Event(`click`,{bubbles:!(r&&t.hasConsumerStoppedPropagation)})),x.current=!1}},[_,s,a]);let T=A.useRef(!O(s)&&s);return(0,j.jsx)(d.input,{type:`checkbox`,"aria-hidden":!0,defaultChecked:l??T.current,required:u,disabled:p,name:m,value:h,form:g,...n,tabIndex:-1,ref:y,onClick:o(t,e=>{x.current&&e.stopPropagation()}),style:{...n.style,...b,position:`absolute`,pointerEvents:`none`,opacity:0,margin:0,transform:`translateX(-100%)`}})},`CheckboxBubbleInput`)),N(D,`isFunction`),N(O,`isIndeterminate`),N(k,`getState`)})))()}function K({className:e,touchTarget:t=!1,...n}){return(0,q.jsx)(B,{"data-slot":`checkbox`,className:y(t?`peer group flex size-11 shrink-0 items-center justify-center rounded-xs outline-none transition-shadow focus-visible:ring-[3px] focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50`:`peer border ${x.borderPrimary} bg-input-background dark:bg-input/30 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground dark:data-[state=checked]:bg-primary data-[state=checked]:border-primary focus-visible:border-ring focus-visible:ring-ring aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive size-4 shrink-0 rounded-xs transition-shadow outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50`,e),...n,children:t?(0,q.jsx)(`span`,{className:`flex size-4 items-center justify-center rounded-xs border ${x.borderPrimary} bg-input-background transition-colors group-data-[state=checked]:border-primary group-data-[state=checked]:bg-primary group-data-[state=checked]:text-primary-foreground dark:bg-input/30 dark:group-data-[state=checked]:bg-primary`,children:(0,q.jsx)(H,{"data-slot":`checkbox-indicator`,className:`flex items-center justify-center text-current transition-none`,children:(0,q.jsx)(C,{className:`size-3.5`})})}):(0,q.jsx)(H,{"data-slot":`checkbox-indicator`,className:`flex items-center justify-center text-current transition-none`,children:(0,q.jsx)(C,{className:`size-3.5`})})},n.checked===void 0?void 0:String(n.checked))}var q;function J(){return(J=t((()=>{n(),G(),S(),b(),v(),q=r(),K.__docgenInfo={description:``,methods:[],displayName:`Checkbox`,props:{touchTarget:{required:!1,tsType:{name:`boolean`},description:``,defaultValue:{value:`false`,computed:!1}}}}})))()}var Y,X,Z,Q,$,te;function ne(){return(ne=t((()=>{J(),w(),Y=r(),X={title:`UI/Checkbox`,component:K,tags:[`autodocs`],argTypes:{checked:{control:`boolean`},touchTarget:{control:`boolean`},disabled:{control:`boolean`}}},Z={render:()=>(0,Y.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,Y.jsx)(K,{id:`cb-default`}),(0,Y.jsx)(T,{htmlFor:`cb-default`,className:`mb-0`,children:`ワクチン済み`})]})},Q={render:()=>(0,Y.jsxs)(`div`,{style:{display:`grid`,gap:12},children:[(0,Y.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,Y.jsx)(K,{id:`cb-off`}),(0,Y.jsx)(T,{htmlFor:`cb-off`,className:`mb-0`,children:`unchecked`})]}),(0,Y.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,Y.jsx)(K,{id:`cb-on`,defaultChecked:!0}),(0,Y.jsx)(T,{htmlFor:`cb-on`,className:`mb-0`,children:`checked`})]}),(0,Y.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,Y.jsx)(K,{id:`cb-disabled`,disabled:!0}),(0,Y.jsx)(T,{htmlFor:`cb-disabled`,className:`mb-0`,children:`disabled`})]}),(0,Y.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,Y.jsx)(K,{id:`cb-invalid`,"aria-invalid":`true`}),(0,Y.jsx)(T,{htmlFor:`cb-invalid`,className:`mb-0`,children:`aria-invalid`})]})]})},$={render:()=>(0,Y.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,Y.jsx)(K,{id:`cb-touch`,touchTarget:!0,defaultChecked:!0}),(0,Y.jsx)(T,{htmlFor:`cb-touch`,className:`mb-0`,children:`タッチ領域 44px（touchTarget）`})]})},te=[`Default`,`States`,`TouchTarget`],Z.parameters={...Z.parameters,docs:{...Z.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    alignItems: "center",
    gap: 8
  }}>
      <Checkbox id="cb-default" />
      <Label htmlFor="cb-default" className="mb-0">
        ワクチン済み
      </Label>
    </div>
}`,...Z.parameters?.docs?.source}}},Q.parameters={...Q.parameters,docs:{...Q.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "grid",
    gap: 12
  }}>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Checkbox id="cb-off" />
        <Label htmlFor="cb-off" className="mb-0">
          unchecked
        </Label>
      </div>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Checkbox id="cb-on" defaultChecked />
        <Label htmlFor="cb-on" className="mb-0">
          checked
        </Label>
      </div>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Checkbox id="cb-disabled" disabled />
        <Label htmlFor="cb-disabled" className="mb-0">
          disabled
        </Label>
      </div>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Checkbox id="cb-invalid" aria-invalid="true" />
        <Label htmlFor="cb-invalid" className="mb-0">
          aria-invalid
        </Label>
      </div>
    </div>
}`,...Q.parameters?.docs?.source}}},$.parameters={...$.parameters,docs:{...$.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    alignItems: "center",
    gap: 8
  }}>
      <Checkbox id="cb-touch" touchTarget defaultChecked />
      <Label htmlFor="cb-touch" className="mb-0">
        タッチ領域 44px（touchTarget）
      </Label>
    </div>
}`,...$.parameters?.docs?.source},description:{story:`touchTarget=true — 44px ヒットエリアラッパー（design-states.md §1）。`,...$.parameters?.docs?.description}}}})))()}ne();export{Z as Default,Q as States,$ as TouchTarget,te as __namedExportsOrder,X as default};
import{a as e,n as t}from"./rolldown-runtime-DkW27tQK.js";import{f as n}from"./iframe-C6Sk8rp0.js";import{t as r}from"./jsx-runtime-DeHZSEgm.js";import{i,n as a,r as o,t as s}from"./dist-CbFlsgBU.js";import{n as c,t as l}from"./dist-S0mBbiUy.js";import{r as u,t as d}from"./dist-oJCNPtvx.js";import{n as f,t as p}from"./dist-DIeBUw0P.js";import{n as m,t as h}from"./dist-BW9ECE_1.js";import{n as g,t as _}from"./utils-DT5hf25b.js";import{n as v,t as y}from"./label-Ct3Dkfk1.js";function b(e){let{__scopeSwitch:t,checked:n,children:r,defaultChecked:i,disabled:a,form:o,name:s,onCheckedChange:c,required:l,value:u=`on`,internal_do_not_use_render:d}=e,[f,p]=m({prop:n,defaultProp:i??!1,onChange:c,caller:D}),[h,g]=C.useState(null),[_,v]=C.useState(null),[y,b]=C.useState(null),S={checked:f,setChecked:p,disabled:a,control:h,setControl:g,name:s,form:o,value:u,userClick:y,onUserClick:b,required:l,defaultChecked:i,isFormControl:!h||!!o||!!h.closest(`form`),bubbleInput:_,setBubbleInput:v};return(0,w.jsx)(A,{scope:t,...S,children:x(d)?d(S):r})}function x(e){return typeof e==`function`}function S(e){return e?`checked`:`unchecked`}var C,w,T,E,D,O,k,A,j,M,N,P,F,I,L,R;function z(){return(z=t((()=>{C=e(n(),1),i(),l(),a(),h(),p(),u(),w=r(),T=Object.defineProperty,E=(e,t)=>T(e,`name`,{value:t,configurable:!0}),D=`Switch`,[O,k]=s(D),[A,j]=O(D),E(b,`SwitchProvider`),M=`SwitchTrigger`,N=C.forwardRef(E(function({__scopeSwitch:e,onClick:t,...n},r){let{control:i,form:a,value:s,disabled:l,checked:u,required:f,setControl:p,setChecked:m,onUserClick:h,isFormControl:g,bubbleInput:_}=j(M,e),v=c(r,p),y=C.useRef(u);return C.useEffect(()=>{let e=a?i?.ownerDocument.getElementById(a):i?.form;if(e instanceof HTMLFormElement){let t=E(()=>m(y.current),`reset`);return e.addEventListener(`reset`,t),()=>e.removeEventListener(`reset`,t)}},[i,a,m]),(0,w.jsx)(d.button,{type:`button`,role:`switch`,"aria-checked":u,"aria-required":f,"data-state":S(u),"data-disabled":l?``:void 0,disabled:l,value:s,...n,ref:v,onClick:o(t,e=>{let t=e.isPropagationStopped();h({requestedChecked:!u,hasConsumerStoppedPropagation:t}),m(e=>!e),_&&g&&(t||e.stopPropagation())})})},`SwitchTrigger`)),P=C.forwardRef(E(function(e,t){let{__scopeSwitch:n,name:r,checked:i,defaultChecked:a,required:o,disabled:s,value:c,onCheckedChange:l,form:u,...d}=e;return(0,w.jsx)(b,{__scopeSwitch:n,checked:i,defaultChecked:a,disabled:s,required:o,onCheckedChange:l,name:r,form:u,value:c,internal_do_not_use_render:({isFormControl:e})=>(0,w.jsxs)(w.Fragment,{children:[(0,w.jsx)(N,{...d,ref:t,__scopeSwitch:n}),e&&(0,w.jsx)(R,{__scopeSwitch:n})]})})},`Switch`)),F=`SwitchThumb`,I=C.forwardRef(E(function(e,t){let{__scopeSwitch:n,...r}=e,i=j(F,n);return(0,w.jsx)(d.span,{"data-state":S(i.checked),"data-disabled":i.disabled?``:void 0,...r,ref:t})},`SwitchThumb`)),L=`SwitchBubbleInput`,R=C.forwardRef(E(function({__scopeSwitch:e,onClick:t,...n},r){let{control:i,userClick:a,checked:s,defaultChecked:l,required:u,disabled:p,name:m,value:h,form:g,bubbleInput:_,setBubbleInput:v}=j(L,e),y=c(r,v),b=f(i),x=C.useRef(!1),S=C.useRef(s),T=C.useRef(a),E=C.useRef(null);C.useEffect(()=>{let e=_;if(!e)return;let t=window.HTMLInputElement.prototype,n=Object.getOwnPropertyDescriptor(t,`checked`).set;if(!n)return;let r=S.current!==s;S.current=s;let i=a!==null&&a!==T.current;if(T.current=a,i)E.current=r?null:a,r&&n.call(e,s),e.dispatchEvent(new Event(`click`,{bubbles:!a.hasConsumerStoppedPropagation}));else if(r){let t=E.current;E.current=null;let r=t?.requestedChecked===s;x.current=!0,n.call(e,s),e.dispatchEvent(new Event(`click`,{bubbles:!(r&&t.hasConsumerStoppedPropagation)})),x.current=!1}},[_,s,a]);let D=C.useRef(s);return(0,w.jsx)(d.input,{type:`checkbox`,"aria-hidden":!0,defaultChecked:l??D.current,required:u,disabled:p,name:m,value:h,form:g,...n,tabIndex:-1,ref:y,onClick:o(t,e=>{x.current&&e.stopPropagation()}),style:{...n.style,...b,position:`absolute`,pointerEvents:`none`,opacity:0,margin:0,transform:`translateX(-100%)`}})},`SwitchBubbleInput`)),E(x,`isFunction`),E(S,`getState`)})))()}function B({className:e,...t}){return(0,V.jsx)(P,{"data-slot":`switch`,className:_(`peer relative inline-flex h-11 w-12 shrink-0 items-center rounded-full border border-transparent bg-transparent transition-all outline-none before:pointer-events-none before:absolute before:inset-x-0 before:top-1/2 before:h-7 before:-translate-y-1/2 before:rounded-full before:bg-switch-background before:transition-colors data-[state=checked]:before:bg-primary dark:data-[state=unchecked]:before:bg-input/80 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50`,e),...t,children:(0,V.jsx)(I,{"data-slot":`switch-thumb`,className:_(`relative z-10 bg-card dark:data-[state=unchecked]:bg-card-foreground dark:data-[state=checked]:bg-primary-foreground pointer-events-none block size-6 rounded-full ring-0 transition-transform data-[state=checked]:translate-x-6 data-[state=unchecked]:translate-x-0`)})},t.checked===void 0?void 0:String(t.checked))}var V;function H(){return(H=t((()=>{n(),z(),g(),V=r(),B.__docgenInfo={description:``,methods:[],displayName:`Switch`}})))()}var U,W,G,K,q;function J(){return(J=t((()=>{H(),v(),U=r(),W={title:`UI/Switch`,component:B,tags:[`autodocs`],argTypes:{checked:{control:`boolean`},disabled:{control:`boolean`}}},G={render:()=>(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-default`}),(0,U.jsx)(y,{htmlFor:`sw-default`,className:`mb-0`,children:`リマインダー送信`})]})},K={render:()=>(0,U.jsxs)(`div`,{style:{display:`grid`,gap:12},children:[(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-on`,defaultChecked:!0}),(0,U.jsx)(y,{htmlFor:`sw-on`,className:`mb-0`,children:`on`})]}),(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-off`}),(0,U.jsx)(y,{htmlFor:`sw-off`,className:`mb-0`,children:`off`})]}),(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-dis`,disabled:!0,defaultChecked:!0}),(0,U.jsx)(y,{htmlFor:`sw-dis`,className:`mb-0`,children:`disabled`})]})]})},q=[`Default`,`States`],G.parameters={...G.parameters,docs:{...G.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    alignItems: "center",
    gap: 8
  }}>
      <Switch id="sw-default" />
      <Label htmlFor="sw-default" className="mb-0">
        リマインダー送信
      </Label>
    </div>
}`,...G.parameters?.docs?.source}}},K.parameters={...K.parameters,docs:{...K.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "grid",
    gap: 12
  }}>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Switch id="sw-on" defaultChecked />
        <Label htmlFor="sw-on" className="mb-0">
          on
        </Label>
      </div>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Switch id="sw-off" />
        <Label htmlFor="sw-off" className="mb-0">
          off
        </Label>
      </div>
      <div style={{
      display: "flex",
      alignItems: "center",
      gap: 8
    }}>
        <Switch id="sw-dis" disabled defaultChecked />
        <Label htmlFor="sw-dis" className="mb-0">
          disabled
        </Label>
      </div>
    </div>
}`,...K.parameters?.docs?.source}}}})))()}J();export{G as Default,K as States,q as __namedExportsOrder,W as default};
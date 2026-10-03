import{i as e,s as t}from"./preload-helper-xPQekRTU.js";import{O as n}from"./iframe-Bt5VGDn5.js";import{t as r}from"./jsx-runtime-CaZkqeYb.js";import{a as i,i as a}from"./dist-BZ7wYAgG.js";import{t as o}from"./utils-DnULZP2N.js";import{t as s}from"./utils-OInfGPY8.js";import{r as c,t as l}from"./dist-DNKdbajt.js";import{n as u,t as d}from"./label-0jG7eM-1.js";import{a as f,i as p,o as m,r as h}from"./dist-DekDGDyr.js";import{n as g,t as _}from"./dist-B4vujon0.js";import{n as v,t as y}from"./dist-CSNypOx3.js";function b(e){let{__scopeSwitch:t,checked:n,children:r,defaultChecked:i,disabled:a,form:o,name:s,onCheckedChange:c,required:l,value:u=`on`,internal_do_not_use_render:d}=e,[f,p]=g({prop:n,defaultProp:i??!1,onChange:c,caller:D}),[m,h]=C.useState(null),[_,v]=C.useState(null),y=C.useRef(!1),[b,S]=C.useReducer(e=>e+1,0),T={checked:f,setChecked:p,disabled:a,control:m,setControl:h,name:s,form:o,value:u,hasConsumerStoppedPropagationRef:y,userInteractionCount:b,onUserInteraction:S,required:l,defaultChecked:i,isFormControl:m?!!o||!!m.closest(`form`):!0,bubbleInput:_,setBubbleInput:v};return(0,w.jsx)(A,{scope:t,...T,children:x(d)?d(T):r})}function x(e){return typeof e==`function`}function S(e){return e?`checked`:`unchecked`}var C,w,T,E,D,O,k,A,j,M,N,P,F,I,L,R,z=e((()=>{C=t(n(),1),m(),a(),p(),_(),y(),c(),w=r(),T=Object.defineProperty,E=(e,t)=>T(e,`name`,{value:t,configurable:!0}),D=`Switch`,[O,k]=h(D),[A,j]=O(D),E(b,`SwitchProvider`),M=`SwitchTrigger`,N=C.forwardRef(E(function({__scopeSwitch:e,onClick:t,...n},r){let{control:a,form:o,value:s,disabled:c,checked:u,required:d,setControl:p,setChecked:m,hasConsumerStoppedPropagationRef:h,onUserInteraction:g,isFormControl:_,bubbleInput:v}=j(M,e),y=i(r,p),b=C.useRef(u);return C.useEffect(()=>{let e=o?a?.ownerDocument.getElementById(o):a?.form;if(e instanceof HTMLFormElement){let t=E(()=>m(b.current),`reset`);return e.addEventListener(`reset`,t),()=>e.removeEventListener(`reset`,t)}},[a,o,m]),(0,w.jsx)(l.button,{type:`button`,role:`switch`,"aria-checked":u,"aria-required":d,"data-state":S(u),"data-disabled":c?``:void 0,disabled:c,value:s,...n,ref:y,onClick:f(t,e=>{g(),m(e=>!e),v&&_&&(h.current=e.isPropagationStopped(),h.current||e.stopPropagation())})})},`SwitchTrigger`)),P=C.forwardRef(E(function(e,t){let{__scopeSwitch:n,name:r,checked:i,defaultChecked:a,required:o,disabled:s,value:c,onCheckedChange:l,form:u,...d}=e;return(0,w.jsx)(b,{__scopeSwitch:n,checked:i,defaultChecked:a,disabled:s,required:o,onCheckedChange:l,name:r,form:u,value:c,internal_do_not_use_render:({isFormControl:e})=>(0,w.jsxs)(w.Fragment,{children:[(0,w.jsx)(N,{...d,ref:t,__scopeSwitch:n}),e&&(0,w.jsx)(R,{__scopeSwitch:n})]})})},`Switch`)),F=`SwitchThumb`,I=C.forwardRef(E(function(e,t){let{__scopeSwitch:n,...r}=e,i=j(F,n);return(0,w.jsx)(l.span,{"data-state":S(i.checked),"data-disabled":i.disabled?``:void 0,...r,ref:t})},`SwitchThumb`)),L=`SwitchBubbleInput`,R=C.forwardRef(E(function({__scopeSwitch:e,onClick:t,...n},r){let{control:a,hasConsumerStoppedPropagationRef:o,userInteractionCount:s,checked:c,defaultChecked:u,required:d,disabled:p,name:m,value:h,form:g,bubbleInput:_,setBubbleInput:y}=j(L,e),b=i(r,y),x=v(a),S=C.useRef(!1),T=C.useRef(c),E=C.useRef(s);C.useEffect(()=>{let e=_;if(!e)return;let t=window.HTMLInputElement.prototype,n=Object.getOwnPropertyDescriptor(t,`checked`).set,r=s!==E.current;E.current=s;let i=T.current!==c;T.current=c;let a=!(r&&o.current);if(i&&n){S.current=!r;let t=new Event(`click`,{bubbles:a});n.call(e,c),e.dispatchEvent(t),S.current=!1}},[_,c,o,s]);let D=C.useRef(c);return(0,w.jsx)(l.input,{type:`checkbox`,"aria-hidden":!0,defaultChecked:u??D.current,required:d,disabled:p,name:m,value:h,form:g,...n,tabIndex:-1,ref:b,onClick:f(t,e=>{S.current&&e.stopPropagation()}),style:{...n.style,...x,position:`absolute`,pointerEvents:`none`,opacity:0,margin:0,transform:`translateX(-100%)`}})},`SwitchBubbleInput`)),E(x,`isFunction`),E(S,`getState`)}));function B({className:e,...t}){return(0,V.jsx)(P,{"data-slot":`switch`,className:o(`peer relative inline-flex h-11 w-12 shrink-0 items-center rounded-full border border-transparent bg-transparent transition-all outline-none before:pointer-events-none before:absolute before:inset-x-0 before:top-1/2 before:h-7 before:-translate-y-1/2 before:rounded-full before:bg-switch-background before:transition-colors data-[state=checked]:before:bg-primary dark:data-[state=unchecked]:before:bg-input/80 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50`,e),...t,children:(0,V.jsx)(I,{"data-slot":`switch-thumb`,className:o(`relative z-10 bg-card dark:data-[state=unchecked]:bg-card-foreground dark:data-[state=checked]:bg-primary-foreground pointer-events-none block size-6 rounded-full ring-0 transition-transform data-[state=checked]:translate-x-6 data-[state=unchecked]:translate-x-0`)})},t.checked===void 0?void 0:String(t.checked))}var V,H=e((()=>{n(),z(),s(),V=r(),B.__docgenInfo={description:``,methods:[],displayName:`Switch`}})),U,W,G,K,q;e((()=>{H(),u(),U=r(),W={title:`UI/Switch`,component:B,tags:[`autodocs`],argTypes:{checked:{control:`boolean`},disabled:{control:`boolean`}}},G={render:()=>(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-default`}),(0,U.jsx)(d,{htmlFor:`sw-default`,className:`mb-0`,children:`リマインダー送信`})]})},K={render:()=>(0,U.jsxs)(`div`,{style:{display:`grid`,gap:12},children:[(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-on`,defaultChecked:!0}),(0,U.jsx)(d,{htmlFor:`sw-on`,className:`mb-0`,children:`on`})]}),(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-off`}),(0,U.jsx)(d,{htmlFor:`sw-off`,className:`mb-0`,children:`off`})]}),(0,U.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,U.jsx)(B,{id:`sw-dis`,disabled:!0,defaultChecked:!0}),(0,U.jsx)(d,{htmlFor:`sw-dis`,className:`mb-0`,children:`disabled`})]})]})},q=[`Default`,`States`],G.parameters={...G.parameters,docs:{...G.parameters?.docs,source:{originalSource:`{
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
}`,...K.parameters?.docs?.source}}}}))();export{G as Default,K as States,q as __namedExportsOrder,W as default};
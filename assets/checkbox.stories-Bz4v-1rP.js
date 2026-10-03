import{i as e,s as t}from"./preload-helper-xPQekRTU.js";import{O as n}from"./iframe-Bt5VGDn5.js";import{t as r}from"./jsx-runtime-CaZkqeYb.js";import{t as i,un as a}from"./lucide-react-yS7Xowms.js";import{a as o,i as s}from"./dist-BZ7wYAgG.js";import{t as c}from"./utils-DnULZP2N.js";import{t as l}from"./utils-OInfGPY8.js";import{r as u,t as d}from"./dist-DNKdbajt.js";import{n as f,t as p}from"./label-0jG7eM-1.js";import{a as m,i as h,o as g,r as _}from"./dist-DekDGDyr.js";import{n as v,t as y}from"./dist-B4vujon0.js";import{n as b,t as x}from"./dist-DLM6oPnB.js";import{n as S,t as C}from"./dist-CSNypOx3.js";function w(e){let{__scopeCheckbox:t,checked:n,children:r,defaultChecked:i,disabled:a,form:o,name:s,onCheckedChange:c,required:l,value:u=`on`,internal_do_not_use_render:d}=e,[f,p]=v({prop:n,defaultProp:i??!1,onChange:c,caller:M}),[m,h]=O.useState(null),[g,_]=O.useState(null),y=O.useRef(!1),[b,x]=O.useReducer(e=>e+1,0),S=m?!!o||!!m.closest(`form`):!0,C={checked:f,disabled:a,setChecked:p,control:m,setControl:h,name:s,form:o,value:u,hasConsumerStoppedPropagationRef:y,userInteractionCount:b,onUserInteraction:x,required:l,defaultChecked:E(i)?!1:i,isFormControl:S,bubbleInput:g,setBubbleInput:_};return(0,k.jsx)(F,{scope:t,...C,children:T(d)?d(C):r})}function T(e){return typeof e==`function`}function E(e){return e===`indeterminate`}function D(e){return E(e)?`indeterminate`:e?`checked`:`unchecked`}var O,k,A,j,M,N,P,F,I,L,R,z,B,V,H,U,W=e((()=>{O=t(n(),1),s(),h(),g(),y(),C(),b(),u(),k=r(),A=Object.defineProperty,j=(e,t)=>A(e,`name`,{value:t,configurable:!0}),M=`Checkbox`,[N,P]=_(M),[F,I]=N(M),j(w,`CheckboxProvider`),L=`CheckboxTrigger`,R=O.forwardRef(j(function({__scopeCheckbox:e,onKeyDown:t,onClick:n,...r},i){let{control:a,value:s,disabled:c,checked:l,required:u,setControl:f,setChecked:p,hasConsumerStoppedPropagationRef:h,onUserInteraction:g,isFormControl:_,bubbleInput:v}=I(L,e),y=o(i,f),b=O.useRef(l);return O.useEffect(()=>{let e=a?.form;if(e){let t=j(()=>p(b.current),`reset`);return e.addEventListener(`reset`,t),()=>e.removeEventListener(`reset`,t)}},[a,p]),(0,k.jsx)(d.button,{type:`button`,role:`checkbox`,"aria-checked":E(l)?`mixed`:l,"aria-required":u,"data-state":D(l),"data-disabled":c?``:void 0,disabled:c,value:s,...r,ref:y,onKeyDown:m(t,e=>{e.key===`Enter`&&e.preventDefault()}),onClick:m(n,e=>{g(),p(e=>E(e)?!0:!e),v&&_&&(h.current=e.isPropagationStopped(),h.current||e.stopPropagation())})})},`CheckboxTrigger`)),z=O.forwardRef(j(function(e,t){let{__scopeCheckbox:n,name:r,checked:i,defaultChecked:a,required:o,disabled:s,value:c,onCheckedChange:l,form:u,...d}=e;return(0,k.jsx)(w,{__scopeCheckbox:n,checked:i,defaultChecked:a,disabled:s,required:o,onCheckedChange:l,name:r,form:u,value:c,internal_do_not_use_render:({isFormControl:e})=>(0,k.jsxs)(k.Fragment,{children:[(0,k.jsx)(R,{...d,ref:t,__scopeCheckbox:n}),e&&(0,k.jsx)(U,{__scopeCheckbox:n})]})})},`Checkbox`)),B=`CheckboxIndicator`,V=O.forwardRef(j(function(e,t){let{__scopeCheckbox:n,forceMount:r,...i}=e,a=I(B,n);return(0,k.jsx)(x,{present:r||E(a.checked)||a.checked===!0,children:(0,k.jsx)(d.span,{"data-state":D(a.checked),"data-disabled":a.disabled?``:void 0,...i,ref:t,style:{pointerEvents:`none`,...e.style}})})},`CheckboxIndicator`)),H=`CheckboxBubbleInput`,U=O.forwardRef(j(function({__scopeCheckbox:e,onClick:t,...n},r){let{control:i,hasConsumerStoppedPropagationRef:a,userInteractionCount:s,checked:c,defaultChecked:l,required:u,disabled:f,name:p,value:h,form:g,bubbleInput:_,setBubbleInput:v}=I(H,e),y=o(r,v),b=S(i),x=O.useRef(!1),C=O.useRef(c),w=O.useRef(s);O.useEffect(()=>{let e=_;if(!e)return;let t=window.HTMLInputElement.prototype,n=Object.getOwnPropertyDescriptor(t,`checked`).set,r=s!==w.current;w.current=s;let i=C.current!==c;C.current=c;let o=!(r&&a.current);if(i&&n){x.current=!r;let t=new Event(`click`,{bubbles:o});e.indeterminate=E(c),n.call(e,E(c)?!1:c),e.dispatchEvent(t),x.current=!1}},[_,c,a,s]);let T=O.useRef(E(c)?!1:c);return(0,k.jsx)(d.input,{type:`checkbox`,"aria-hidden":!0,defaultChecked:l??T.current,required:u,disabled:f,name:p,value:h,form:g,...n,tabIndex:-1,ref:y,onClick:m(t,e=>{x.current&&e.stopPropagation()}),style:{...n.style,...b,position:`absolute`,pointerEvents:`none`,opacity:0,margin:0,transform:`translateX(-100%)`}})},`CheckboxBubbleInput`)),j(T,`isFunction`),j(E,`isIndeterminate`),j(D,`getState`)}));function G({className:e,touchTarget:t=!1,...n}){return(0,K.jsx)(z,{"data-slot":`checkbox`,className:c(t?`peer group flex size-11 shrink-0 items-center justify-center rounded-xs outline-none transition-shadow focus-visible:ring-[3px] focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50`:`peer border bg-input-background dark:bg-input/30 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground dark:data-[state=checked]:bg-primary data-[state=checked]:border-primary focus-visible:border-ring focus-visible:ring-ring aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive size-4 shrink-0 rounded-xs border transition-shadow outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50`,e),...n,children:t?(0,K.jsx)(`span`,{className:`flex size-4 items-center justify-center rounded-xs border bg-input-background transition-colors group-data-[state=checked]:border-primary group-data-[state=checked]:bg-primary group-data-[state=checked]:text-primary-foreground dark:bg-input/30 dark:group-data-[state=checked]:bg-primary`,children:(0,K.jsx)(V,{"data-slot":`checkbox-indicator`,className:`flex items-center justify-center text-current transition-none`,children:(0,K.jsx)(a,{className:`size-3.5`})})}):(0,K.jsx)(V,{"data-slot":`checkbox-indicator`,className:`flex items-center justify-center text-current transition-none`,children:(0,K.jsx)(a,{className:`size-3.5`})})},n.checked===void 0?void 0:String(n.checked))}var K,q=e((()=>{n(),W(),i(),l(),K=r(),G.__docgenInfo={description:``,methods:[],displayName:`Checkbox`,props:{touchTarget:{required:!1,tsType:{name:`boolean`},description:``,defaultValue:{value:`false`,computed:!1}}}}})),J,Y,X,Z,Q,$;e((()=>{q(),f(),J=r(),Y={title:`UI/Checkbox`,component:G,tags:[`autodocs`],argTypes:{checked:{control:`boolean`},touchTarget:{control:`boolean`},disabled:{control:`boolean`}}},X={render:()=>(0,J.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,J.jsx)(G,{id:`cb-default`}),(0,J.jsx)(p,{htmlFor:`cb-default`,className:`mb-0`,children:`ワクチン済み`})]})},Z={render:()=>(0,J.jsxs)(`div`,{style:{display:`grid`,gap:12},children:[(0,J.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,J.jsx)(G,{id:`cb-off`}),(0,J.jsx)(p,{htmlFor:`cb-off`,className:`mb-0`,children:`unchecked`})]}),(0,J.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,J.jsx)(G,{id:`cb-on`,defaultChecked:!0}),(0,J.jsx)(p,{htmlFor:`cb-on`,className:`mb-0`,children:`checked`})]}),(0,J.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,J.jsx)(G,{id:`cb-disabled`,disabled:!0}),(0,J.jsx)(p,{htmlFor:`cb-disabled`,className:`mb-0`,children:`disabled`})]}),(0,J.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,J.jsx)(G,{id:`cb-invalid`,"aria-invalid":`true`}),(0,J.jsx)(p,{htmlFor:`cb-invalid`,className:`mb-0`,children:`aria-invalid`})]})]})},Q={render:()=>(0,J.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:8},children:[(0,J.jsx)(G,{id:`cb-touch`,touchTarget:!0,defaultChecked:!0}),(0,J.jsx)(p,{htmlFor:`cb-touch`,className:`mb-0`,children:`タッチ領域 44px（touchTarget）`})]})},$=[`Default`,`States`,`TouchTarget`],X.parameters={...X.parameters,docs:{...X.parameters?.docs,source:{originalSource:`{
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
}`,...X.parameters?.docs?.source}}},Z.parameters={...Z.parameters,docs:{...Z.parameters?.docs,source:{originalSource:`{
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
}`,...Z.parameters?.docs?.source}}},Q.parameters={...Q.parameters,docs:{...Q.parameters?.docs,source:{originalSource:`{
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
}`,...Q.parameters?.docs?.source},description:{story:`touchTarget=true — 44px ヒットエリアラッパー（design-states.md §1）。`,...Q.parameters?.docs?.description}}}}))();export{X as Default,Z as States,Q as TouchTarget,$ as __namedExportsOrder,Y as default};
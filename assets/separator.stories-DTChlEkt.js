import{i as e,s as t}from"./preload-helper-xPQekRTU.js";import{O as n}from"./iframe-Bt5VGDn5.js";import{t as r}from"./jsx-runtime-CaZkqeYb.js";import{t as i}from"./utils-DnULZP2N.js";import{t as a}from"./utils-OInfGPY8.js";import{r as o,t as s}from"./dist-DNKdbajt.js";function c(e){return m.includes(e)}var l,u,d,f,p,m,h,g,_=e((()=>{l=t(n(),1),o(),u=r(),d=Object.defineProperty,f=(e,t)=>d(e,`name`,{value:t,configurable:!0}),p=`horizontal`,m=[`horizontal`,`vertical`],h=l.forwardRef(f(function(e,t){let{decorative:n,orientation:r=p,...i}=e,a=c(r)?r:p,o=n?{role:`none`}:{"aria-orientation":a===`vertical`?a:void 0,role:`separator`};return(0,u.jsx)(s.div,{"data-orientation":a,...o,...i,ref:t})},`Separator`)),f(c,`isValidOrientation`),g=h}));function v({className:e,orientation:t=`horizontal`,decorative:n=!0,...r}){return(0,y.jsx)(g,{"data-slot":`separator-root`,decorative:n,orientation:t,className:i(`bg-border shrink-0 data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-full data-[orientation=vertical]:w-px`,e),...r})}var y,b=e((()=>{n(),_(),a(),y=r(),v.__docgenInfo={description:``,methods:[],displayName:`Separator`,props:{orientation:{defaultValue:{value:`"horizontal"`,computed:!1},required:!1},decorative:{defaultValue:{value:`true`,computed:!1},required:!1}}}})),x,S,C,w,T;e((()=>{b(),x=r(),S={title:`UI/Separator`,component:v,tags:[`autodocs`]},C={render:()=>(0,x.jsxs)(`div`,{style:{maxWidth:320},children:[(0,x.jsx)(`p`,{className:`text-sm`,children:`上のセクション`}),(0,x.jsx)(v,{className:`my-4`}),(0,x.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`下のセクション`})]})},w={render:()=>(0,x.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:16,height:40},children:[(0,x.jsx)(`span`,{className:`text-sm`,children:`編集`}),(0,x.jsx)(v,{orientation:`vertical`}),(0,x.jsx)(`span`,{className:`text-sm`,children:`印刷`}),(0,x.jsx)(v,{orientation:`vertical`}),(0,x.jsx)(`span`,{className:`text-sm text-destructive`,children:`削除`})]})},T=[`Horizontal`,`Vertical`],C.parameters={...C.parameters,docs:{...C.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    maxWidth: 320
  }}>
      <p className="text-sm">上のセクション</p>
      <Separator className="my-4" />
      <p className="text-sm text-muted-foreground">下のセクション</p>
    </div>
}`,...C.parameters?.docs?.source}}},w.parameters={...w.parameters,docs:{...w.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    alignItems: "center",
    gap: 16,
    height: 40
  }}>
      <span className="text-sm">編集</span>
      <Separator orientation="vertical" />
      <span className="text-sm">印刷</span>
      <Separator orientation="vertical" />
      <span className="text-sm text-destructive">削除</span>
    </div>
}`,...w.parameters?.docs?.source}}}}))();export{C as Horizontal,w as Vertical,T as __namedExportsOrder,S as default};
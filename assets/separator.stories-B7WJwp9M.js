import{a as e,n as t}from"./rolldown-runtime-DkW27tQK.js";import{f as n}from"./iframe-C6Sk8rp0.js";import{t as r}from"./jsx-runtime-DeHZSEgm.js";import{r as i,t as a}from"./dist-oJCNPtvx.js";import{n as o,t as s}from"./utils-DT5hf25b.js";function c(e){return h.includes(e)}var l,u,d,f,p,m,h,g;function _(){return(_=t((()=>{l=e(n(),1),i(),u=r(),d=Object.defineProperty,f=(e,t)=>d(e,`name`,{value:t,configurable:!0}),p={Horizontal:`horizontal`,Vertical:`vertical`},m=p.Horizontal,h=Object.values(p),g=l.forwardRef(f(function(e,t){let{decorative:n,orientation:r=m,...i}=e,o=c(r)?r:m,s=o===p.Vertical?o:void 0,l=n?{role:`none`}:{"aria-orientation":s,role:`separator`};return(0,u.jsx)(a.div,{"data-orientation":o,...l,...i,ref:t})},`Separator`)),f(c,`isValidOrientation`)})))()}function v({className:e,orientation:t=`horizontal`,decorative:n=!0,...r}){return(0,y.jsx)(g,{"data-slot":`separator-root`,decorative:n,orientation:t,className:s(`bg-border shrink-0 data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-full data-[orientation=vertical]:w-px`,e),...r})}var y;function b(){return(b=t((()=>{n(),_(),o(),y=r(),v.__docgenInfo={description:``,methods:[],displayName:`Separator`,props:{orientation:{defaultValue:{value:`"horizontal"`,computed:!1},required:!1},decorative:{defaultValue:{value:`true`,computed:!1},required:!1}}}})))()}var x,S,C,w,T;function E(){return(E=t((()=>{b(),x=r(),S={title:`UI/Separator`,component:v,tags:[`autodocs`]},C={render:()=>(0,x.jsxs)(`div`,{style:{maxWidth:320},children:[(0,x.jsx)(`p`,{className:`text-sm`,children:`上のセクション`}),(0,x.jsx)(v,{className:`my-4`}),(0,x.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`下のセクション`})]})},w={render:()=>(0,x.jsxs)(`div`,{style:{display:`flex`,alignItems:`center`,gap:16,height:40},children:[(0,x.jsx)(`span`,{className:`text-sm`,children:`編集`}),(0,x.jsx)(v,{orientation:`vertical`}),(0,x.jsx)(`span`,{className:`text-sm`,children:`印刷`}),(0,x.jsx)(v,{orientation:`vertical`}),(0,x.jsx)(`span`,{className:`text-sm text-destructive`,children:`削除`})]})},T=[`Horizontal`,`Vertical`],C.parameters={...C.parameters,docs:{...C.parameters?.docs,source:{originalSource:`{
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
}`,...w.parameters?.docs?.source}}}})))()}E();export{C as Horizontal,w as Vertical,T as __namedExportsOrder,S as default};
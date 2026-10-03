import{i as e,s as t}from"./preload-helper-xPQekRTU.js";import{O as n}from"./iframe-Bt5VGDn5.js";import{t as r}from"./react-dom-DDCpVvBt.js";import{t as i}from"./jsx-runtime-CaZkqeYb.js";import{d as a,l as o}from"./design-tokens-CoIxcO1x.js";function s({testId:e,orientation:t=`portrait`,active:n=!0,children:r}){let i=(0,c.useRef)(null);return(0,c.useEffect)(()=>{},[]),typeof document>`u`?null:(0,l.createPortal)((0,u.jsxs)(`div`,{ref:i,className:`hidden print:block bg-white`,"data-testid":e,"data-print-portal":``,"data-print-active":n?`true`:`false`,style:{position:`fixed`,inset:0,zIndex:o.overlay,overflow:`auto`,padding:`8mm`},children:[(0,u.jsx)(`style`,{type:`text/css`,children:`
          @media print {
            /* アプリ本体・トースト等（印刷ポータル以外の body 直下要素）を隠す。
               除外キーは全ポータル共通の固定属性 data-print-portal。per-instance な
               testId に依存しないため、複数ポータル同居時も互いを隠さない（#187）。*/
            body > :not([data-print-portal]) { display: none !important; }
            /* 印刷ポータルは既定で表示。active 未指定（既定）の同居では全ポータルが残り白紙化しない。
               全ポータルを active=false にした場合のみ印刷面が 0 になり得る（呼び出し側で 1 面を active に保つ）。*/
            [data-print-portal] { display: block !important; }
            /* 明示的に非アクティブ指定されたポータルのみ隠す
               （同居時に印刷面を 1 件へ限定する手段。specificity を 1 段上げて表示ルールに勝たせる）。*/
            [data-print-portal][data-print-active="false"] { display: none !important; }
            body { margin: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
          }
          @page { size: A4 ${t===`landscape`?`landscape`:`portrait`}; margin: 8mm; }
        `}),r]}),document.body)}var c,l,u,d=e((()=>{c=t(n(),1),l=t(r(),1),a(),u=i()})),f,p,m,h,g;e((()=>{d(),f=i(),p={title:`Shared/PrintPortal`,component:s,tags:[`autodocs`],args:{testId:`print-demo`,orientation:`portrait`,children:(0,f.jsxs)(`div`,{children:[(0,f.jsx)(`h1`,{children:`印刷用レイアウト（A4 portrait）`}),(0,f.jsx)(`p`,{children:`この内容は document.body 直下のポータルに描画されます。`})]})}},m={},h={args:{testId:`print-demo-landscape`,orientation:`landscape`}},g=[`Portrait`,`Landscape`],m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  args: {
    testId: "print-demo-landscape",
    orientation: "landscape"
  }
}`,...h.parameters?.docs?.source}}}}))();export{h as Landscape,m as Portrait,g as __namedExportsOrder,p as default};
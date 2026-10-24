import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{f as t}from"./iframe-C6Sk8rp0.js";import{t as n}from"./react-dom-B_T95ef7.js";import{t as r}from"./jsx-runtime-DeHZSEgm.js";import{d as i,l as a}from"./design-tokens-B-kSoC24.js";function o({testId:e,orientation:t=`portrait`,active:n=!0,children:r}){let i=(0,s.useRef)(null);return(0,s.useEffect)(()=>{},[]),typeof document>`u`?null:(0,c.createPortal)((0,l.jsxs)(`div`,{ref:i,className:`hidden print:block bg-white`,"data-testid":e,"data-print-portal":``,"data-print-active":n?`true`:`false`,style:{position:`fixed`,inset:0,zIndex:a.overlay,overflow:`auto`,padding:`8mm`},children:[(0,l.jsx)(`style`,{type:`text/css`,children:`
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
        `}),r]}),document.body)}var s,c,l;function u(){return(u=e((()=>{s=t(),c=n(),i(),l=r()})))()}var d,f,p,m,h;function g(){return(g=e((()=>{u(),d=r(),f={title:`Shared/PrintPortal`,component:o,tags:[`autodocs`],args:{testId:`print-demo`,orientation:`portrait`,children:(0,d.jsxs)(`div`,{children:[(0,d.jsx)(`h1`,{children:`印刷用レイアウト（A4 portrait）`}),(0,d.jsx)(`p`,{children:`この内容は document.body 直下のポータルに描画されます。`})]})}},p={},m={args:{testId:`print-demo-landscape`,orientation:`landscape`}},h=[`Portrait`,`Landscape`],p.parameters={...p.parameters,docs:{...p.parameters?.docs,source:{originalSource:`{}`,...p.parameters?.docs?.source}}},m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  args: {
    testId: "print-demo-landscape",
    orientation: "landscape"
  }
}`,...m.parameters?.docs?.source}}}})))()}g();export{m as Landscape,p as Portrait,h as __namedExportsOrder,f as default};
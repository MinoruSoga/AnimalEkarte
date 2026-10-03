import{i as e}from"./preload-helper-xPQekRTU.js";import{O as t}from"./iframe-Bt5VGDn5.js";import{t as n}from"./jsx-runtime-CaZkqeYb.js";import{t as r}from"./utils-DnULZP2N.js";import{t as i}from"./utils-OInfGPY8.js";import{n as a,t as o}from"./button-DzIL5X86.js";function s({className:e,ref:t,...n}){return(0,d.jsx)(`div`,{ref:t,"data-slot":`card`,className:r(`bg-card text-card-foreground flex flex-col gap-6 rounded-lg border`,e),...n})}function c({className:e,ref:t,...n}){return(0,d.jsx)(`div`,{ref:t,"data-slot":`card-header`,className:r(`@container/card-header grid auto-rows-min grid-rows-[auto_auto] items-start gap-1.5 px-6 pt-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6`,e),...n})}function l({className:e,ref:t,...n}){return(0,d.jsx)(`h4`,{ref:t,"data-slot":`card-title`,className:r(`leading-none`,e),...n})}function u({className:e,ref:t,...n}){return(0,d.jsx)(`div`,{ref:t,"data-slot":`card-content`,className:r(`px-6 [&:last-child]:pb-6`,e),...n})}var d,f=e((()=>{t(),i(),d=n(),s.__docgenInfo={description:``,methods:[],displayName:`Card`,props:{ref:{required:!1,tsType:{name:`ReactRef`,raw:`React.Ref<HTMLDivElement>`,elements:[{name:`HTMLDivElement`}]},description:``}}},c.__docgenInfo={description:``,methods:[],displayName:`CardHeader`,props:{ref:{required:!1,tsType:{name:`ReactRef`,raw:`React.Ref<HTMLDivElement>`,elements:[{name:`HTMLDivElement`}]},description:``}}},l.__docgenInfo={description:``,methods:[],displayName:`CardTitle`,props:{ref:{required:!1,tsType:{name:`ReactRef`,raw:`React.Ref<HTMLParagraphElement>`,elements:[{name:`HTMLParagraphElement`}]},description:``}}},u.__docgenInfo={description:``,methods:[],displayName:`CardContent`,props:{ref:{required:!1,tsType:{name:`ReactRef`,raw:`React.Ref<HTMLDivElement>`,elements:[{name:`HTMLDivElement`}]},description:``}}}})),p,m,h,g,_;e((()=>{f(),a(),p=n(),m={title:`UI/Card`,component:s,tags:[`autodocs`]},h={render:()=>(0,p.jsxs)(s,{style:{maxWidth:360},children:[(0,p.jsx)(c,{children:(0,p.jsx)(l,{children:`本日の予約`})}),(0,p.jsx)(u,{children:(0,p.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`午前 9 件 / 午後 12 件`})})]})},g={render:()=>(0,p.jsxs)(s,{style:{maxWidth:360},children:[(0,p.jsxs)(c,{children:[(0,p.jsx)(l,{children:`会計待ち`}),(0,p.jsx)(o,{variant:`outline`,size:`sm`,"data-slot":`card-action`,children:`すべて表示`})]}),(0,p.jsx)(u,{children:(0,p.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`3 件`})})]})},_=[`Default`,`WithAction`],h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <Card style={{
    maxWidth: 360
  }}>
      <CardHeader>
        <CardTitle>本日の予約</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">午前 9 件 / 午後 12 件</p>
      </CardContent>
    </Card>
}`,...h.parameters?.docs?.source}}},g.parameters={...g.parameters,docs:{...g.parameters?.docs,source:{originalSource:`{
  render: () => <Card style={{
    maxWidth: 360
  }}>
      <CardHeader>
        <CardTitle>会計待ち</CardTitle>
        <Button variant="outline" size="sm" data-slot="card-action">
          すべて表示
        </Button>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">3 件</p>
      </CardContent>
    </Card>
}`,...g.parameters?.docs?.source}}}}))();export{h as Default,g as WithAction,_ as __namedExportsOrder,m as default};
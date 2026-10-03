import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{a as n,i as r,n as i,r as a,t as o}from"./tabs-DUPlE5QT.js";var s,c,l,u,d;e((()=>{n(),s=t(),c={title:`UI/Tabs`,component:o,tags:[`autodocs`]},l={render:()=>(0,s.jsxs)(o,{defaultValue:`record`,style:{maxWidth:480},children:[(0,s.jsxs)(a,{children:[(0,s.jsx)(r,{value:`record`,children:`カルテ`}),(0,s.jsx)(r,{value:`vaccine`,children:`ワクチン`}),(0,s.jsx)(r,{value:`lab`,children:`検査`})]}),(0,s.jsx)(i,{value:`record`,children:(0,s.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`カルテ一覧のコンテンツ`})}),(0,s.jsx)(i,{value:`vaccine`,children:(0,s.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`ワクチン履歴のコンテンツ`})}),(0,s.jsx)(i,{value:`lab`,children:(0,s.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`検査結果のコンテンツ`})})]})},u={render:()=>(0,s.jsxs)(o,{defaultValue:`record`,style:{maxWidth:480},children:[(0,s.jsxs)(a,{children:[(0,s.jsx)(r,{value:`record`,children:`カルテ`}),(0,s.jsx)(r,{value:`archive`,disabled:!0,children:`過去カルテ（権限なし）`})]}),(0,s.jsx)(i,{value:`record`,children:(0,s.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`権限のないタブは disabled — 視覚だけでなく操作も不可。`})})]})},d=[`Default`,`WithDisabled`],l.parameters={...l.parameters,docs:{...l.parameters?.docs,source:{originalSource:`{
  render: () => <Tabs defaultValue="record" style={{
    maxWidth: 480
  }}>
      <TabsList>
        <TabsTrigger value="record">カルテ</TabsTrigger>
        <TabsTrigger value="vaccine">ワクチン</TabsTrigger>
        <TabsTrigger value="lab">検査</TabsTrigger>
      </TabsList>
      <TabsContent value="record">
        <p className="text-sm text-muted-foreground">カルテ一覧のコンテンツ</p>
      </TabsContent>
      <TabsContent value="vaccine">
        <p className="text-sm text-muted-foreground">ワクチン履歴のコンテンツ</p>
      </TabsContent>
      <TabsContent value="lab">
        <p className="text-sm text-muted-foreground">検査結果のコンテンツ</p>
      </TabsContent>
    </Tabs>
}`,...l.parameters?.docs?.source}}},u.parameters={...u.parameters,docs:{...u.parameters?.docs,source:{originalSource:`{
  render: () => <Tabs defaultValue="record" style={{
    maxWidth: 480
  }}>
      <TabsList>
        <TabsTrigger value="record">カルテ</TabsTrigger>
        <TabsTrigger value="archive" disabled>
          過去カルテ（権限なし）
        </TabsTrigger>
      </TabsList>
      <TabsContent value="record">
        <p className="text-sm text-muted-foreground">
          権限のないタブは disabled — 視覚だけでなく操作も不可。
        </p>
      </TabsContent>
    </Tabs>
}`,...u.parameters?.docs?.source}}}}))();export{l as Default,u as WithDisabled,d as __namedExportsOrder,c as default};
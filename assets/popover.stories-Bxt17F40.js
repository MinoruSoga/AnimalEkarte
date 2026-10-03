import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./button-DzIL5X86.js";import{n as i,t as a}from"./input-D5MfHkT8.js";import{n as o,t as s}from"./label-0jG7eM-1.js";import{a as c,i as l,r as u,t as d}from"./popover-D6KyWpBK.js";var f,p,m,h,g;e((()=>{c(),n(),i(),o(),f=t(),p={title:`UI/Popover`,component:d,tags:[`autodocs`]},m={render:()=>(0,f.jsxs)(d,{children:[(0,f.jsx)(l,{asChild:!0,children:(0,f.jsx)(r,{variant:`outline`,children:`フィルタ`})}),(0,f.jsx)(u,{children:(0,f.jsxs)(`div`,{style:{display:`grid`,gap:8},children:[(0,f.jsx)(s,{htmlFor:`pop-filter`,children:`表示件数`}),(0,f.jsx)(a,{id:`pop-filter`,defaultValue:`20`})]})})]})},h={render:()=>(0,f.jsxs)(d,{defaultOpen:!0,children:[(0,f.jsx)(l,{asChild:!0,children:(0,f.jsx)(r,{variant:`outline`,children:`フィルタ`})}),(0,f.jsx)(u,{children:(0,f.jsx)(`p`,{className:`text-sm text-muted-foreground`,children:`Popover コンテンツ`})})]})},g=[`Default`,`Open`],m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  render: () => <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">フィルタ</Button>
      </PopoverTrigger>
      <PopoverContent>
        <div style={{
        display: "grid",
        gap: 8
      }}>
          <Label htmlFor="pop-filter">表示件数</Label>
          <Input id="pop-filter" defaultValue="20" />
        </div>
      </PopoverContent>
    </Popover>
}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <Popover defaultOpen>
      <PopoverTrigger asChild>
        <Button variant="outline">フィルタ</Button>
      </PopoverTrigger>
      <PopoverContent>
        <p className="text-sm text-muted-foreground">Popover コンテンツ</p>
      </PopoverContent>
    </Popover>
}`,...h.parameters?.docs?.source}}}}))();export{m as Default,h as Open,g as __namedExportsOrder,p as default};
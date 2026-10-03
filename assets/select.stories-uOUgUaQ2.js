import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./label-0jG7eM-1.js";import{a as i,i as a,n as o,o as s,r as c,t as l}from"./select-QiJ9CgwX.js";var u,d,f,p,m,h,g,_;e((()=>{s(),n(),u=t(),d={title:`UI/Select`,component:l,tags:[`autodocs`]},f=[{value:`dog`,label:`犬`},{value:`cat`,label:`猫`},{value:`rabbit`,label:`うさぎ`},{value:`other`,label:`その他`}],p={render:()=>(0,u.jsx)(`div`,{style:{maxWidth:320},children:(0,u.jsxs)(l,{children:[(0,u.jsx)(a,{children:(0,u.jsx)(i,{placeholder:`種別を選択`})}),(0,u.jsx)(o,{children:f.map(e=>(0,u.jsx)(c,{value:e.value,children:e.label},e.value))})]})})},m={render:()=>(0,u.jsxs)(`div`,{style:{display:`grid`,gap:8,maxWidth:320},children:[(0,u.jsx)(r,{htmlFor:`species`,children:`動物種別`}),(0,u.jsxs)(l,{defaultValue:`cat`,children:[(0,u.jsx)(a,{id:`species`,children:(0,u.jsx)(i,{})}),(0,u.jsx)(o,{children:f.map(e=>(0,u.jsx)(c,{value:e.value,children:e.label},e.value))})]})]})},h={render:()=>(0,u.jsx)(`div`,{style:{maxWidth:320},children:(0,u.jsxs)(l,{children:[(0,u.jsx)(a,{"aria-invalid":`true`,children:(0,u.jsx)(i,{placeholder:`未選択（必須）`})}),(0,u.jsx)(o,{children:f.map(e=>(0,u.jsx)(c,{value:e.value,children:e.label},e.value))})]})})},g={render:()=>(0,u.jsx)(`div`,{style:{maxWidth:320},children:(0,u.jsxs)(l,{disabled:!0,defaultValue:`dog`,children:[(0,u.jsx)(a,{children:(0,u.jsx)(i,{})}),(0,u.jsx)(o,{children:f.map(e=>(0,u.jsx)(c,{value:e.value,children:e.label},e.value))})]})})},_=[`Default`,`WithLabel`,`Invalid`,`Disabled`],p.parameters={...p.parameters,docs:{...p.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    maxWidth: 320
  }}>
      <Select>
        <SelectTrigger>
          <SelectValue placeholder="種別を選択" />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map(s => <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>)}
        </SelectContent>
      </Select>
    </div>
}`,...p.parameters?.docs?.source}}},m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "grid",
    gap: 8,
    maxWidth: 320
  }}>
      <Label htmlFor="species">動物種別</Label>
      <Select defaultValue="cat">
        <SelectTrigger id="species">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map(s => <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>)}
        </SelectContent>
      </Select>
    </div>
}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    maxWidth: 320
  }}>
      <Select>
        <SelectTrigger aria-invalid="true">
          <SelectValue placeholder="未選択（必須）" />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map(s => <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>)}
        </SelectContent>
      </Select>
    </div>
}`,...h.parameters?.docs?.source}}},g.parameters={...g.parameters,docs:{...g.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    maxWidth: 320
  }}>
      <Select disabled defaultValue="dog">
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {SPECIES.map(s => <SelectItem key={s.value} value={s.value}>
              {s.label}
            </SelectItem>)}
        </SelectContent>
      </Select>
    </div>
}`,...g.parameters?.docs?.source}}}}))();export{p as Default,g as Disabled,h as Invalid,m as WithLabel,_ as __namedExportsOrder,d as default};
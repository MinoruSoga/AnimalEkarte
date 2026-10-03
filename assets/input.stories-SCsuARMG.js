import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./input-D5MfHkT8.js";import{n as i,t as a}from"./label-0jG7eM-1.js";var o,s,c,l,u,d,f;e((()=>{n(),i(),o=t(),s={title:`UI/Input`,component:r,tags:[`autodocs`],argTypes:{disabled:{control:`boolean`},placeholder:{control:`text`}},args:{placeholder:`飼主名を入力`},parameters:{docs:{description:{component:`design-states.md §3 の状態（hover/focus/disabled/invalid）を確認するカタログ。`}}}},c={},l={render:()=>(0,o.jsxs)(`div`,{style:{display:`grid`,gap:8,maxWidth:320},children:[(0,o.jsx)(a,{htmlFor:`owner-name`,children:`飼主名`}),(0,o.jsx)(r,{id:`owner-name`,placeholder:`例: 山田 太郎`})]})},u={render:()=>(0,o.jsxs)(`div`,{style:{display:`grid`,gap:8,maxWidth:320},children:[(0,o.jsx)(a,{htmlFor:`invalid-input`,children:`電話番号`}),(0,o.jsx)(r,{id:`invalid-input`,"aria-invalid":`true`,defaultValue:`090-`}),(0,o.jsx)(`p`,{className:`text-sm text-destructive`,children:`ハイフンを含めて入力してください`})]})},d={args:{disabled:!0,defaultValue:`読み取り専用`}},f=[`Default`,`WithLabel`,`Invalid`,`Disabled`],c.parameters={...c.parameters,docs:{...c.parameters?.docs,source:{originalSource:`{}`,...c.parameters?.docs?.source}}},l.parameters={...l.parameters,docs:{...l.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "grid",
    gap: 8,
    maxWidth: 320
  }}>
      <Label htmlFor="owner-name">飼主名</Label>
      <Input id="owner-name" placeholder="例: 山田 太郎" />
    </div>
}`,...l.parameters?.docs?.source}}},u.parameters={...u.parameters,docs:{...u.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "grid",
    gap: 8,
    maxWidth: 320
  }}>
      <Label htmlFor="invalid-input">電話番号</Label>
      <Input id="invalid-input" aria-invalid="true" defaultValue="090-" />
      <p className="text-sm text-destructive">ハイフンを含めて入力してください</p>
    </div>
}`,...u.parameters?.docs?.source},description:{story:`aria-invalid=true で destructive border + bg/5（design-states.md §3 invalid）。`,...u.parameters?.docs?.description}}},d.parameters={...d.parameters,docs:{...d.parameters?.docs,source:{originalSource:`{
  args: {
    disabled: true,
    defaultValue: "読み取り専用"
  }
}`,...d.parameters?.docs?.source}}}}))();export{c as Default,d as Disabled,u as Invalid,l as WithLabel,f as __namedExportsOrder,s as default};
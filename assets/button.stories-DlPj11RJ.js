import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{t}from"./jsx-runtime-DeHZSEgm.js";import{n,t as r}from"./button-CKac4K-h.js";import{n as i,t as a}from"./plus-Ba1lhQ3a.js";import{n as o,t as s}from"./trash-DV-uHq1B.js";var c,l,u,d,f,p,m,h,g;function _(){return(_=e((()=>{i(),o(),n(),c=t(),l={title:`UI/Button`,component:r,tags:[`autodocs`],argTypes:{variant:{control:`select`,options:[`default`,`primary`,`destructive`,`outline`,`secondary`,`ghost`,`link`,`ghost-danger`]},size:{control:`select`,options:[`default`,`sm`,`lg`,`icon`]},loading:{control:`boolean`},disabled:{control:`boolean`}},args:{children:`保存`},parameters:{docs:{description:{component:`design-system.md §7.2 + design-states.md §1–3 の仕様を全状態で確認するカタログ。
focus-visible ring はキーボード起因のみ発火するため、Storybook 上の検証は
\`button.test.tsx\`（EMR-207 回帰）と axe（addon-a11y）で担保する。`}}}},u={},d={render:()=>(0,c.jsxs)(`div`,{style:{display:`flex`,gap:12,flexWrap:`wrap`,alignItems:`center`},children:[(0,c.jsx)(r,{variant:`default`,children:`primary（既定）`}),(0,c.jsx)(r,{variant:`secondary`,children:`secondary`}),(0,c.jsx)(r,{variant:`outline`,children:`outline`}),(0,c.jsx)(r,{variant:`ghost`,children:`ghost`}),(0,c.jsx)(r,{variant:`link`,children:`link`}),(0,c.jsx)(r,{variant:`destructive`,children:`destructive`}),(0,c.jsx)(r,{variant:`ghost-danger`,children:`ghost-danger`})]})},f={render:()=>(0,c.jsxs)(`div`,{style:{display:`flex`,gap:12,alignItems:`center`},children:[(0,c.jsx)(r,{size:`sm`,children:`sm`}),(0,c.jsx)(r,{size:`default`,children:`default`}),(0,c.jsx)(r,{size:`lg`,children:`lg`}),(0,c.jsx)(r,{size:`icon`,"aria-label":`追加`,children:(0,c.jsx)(a,{})})]})},p={args:{loading:!0,children:`保存中`}},m={args:{disabled:!0,children:`保存`}},h={render:()=>(0,c.jsxs)(`div`,{style:{display:`flex`,gap:8},children:[(0,c.jsxs)(r,{variant:`ghost`,className:`h-9 px-3 text-sm`,children:[(0,c.jsx)(s,{}),` 削除`]}),(0,c.jsx)(r,{variant:`outline`,className:`h-9 px-3 text-sm`,children:`編集`})]})},g=[`Default`,`AllVariants`,`Sizes`,`Loading`,`Disabled`,`DenseRowAction`],u.parameters={...u.parameters,docs:{...u.parameters?.docs,source:{originalSource:`{}`,...u.parameters?.docs?.source},description:{story:`Primary CTA — pill + #038B94。design-system.md §7.2 button-primary。`,...u.parameters?.docs?.description}}},d.parameters={...d.parameters,docs:{...d.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    gap: 12,
    flexWrap: "wrap",
    alignItems: "center"
  }}>
      <Button variant="default">primary（既定）</Button>
      <Button variant="secondary">secondary</Button>
      <Button variant="outline">outline</Button>
      <Button variant="ghost">ghost</Button>
      <Button variant="link">link</Button>
      <Button variant="destructive">destructive</Button>
      <Button variant="ghost-danger">ghost-danger</Button>
    </div>
}`,...d.parameters?.docs?.source}}},f.parameters={...f.parameters,docs:{...f.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    gap: 12,
    alignItems: "center"
  }}>
      <Button size="sm">sm</Button>
      <Button size="default">default</Button>
      <Button size="lg">lg</Button>
      <Button size="icon" aria-label="追加">
        <Plus />
      </Button>
    </div>
}`,...f.parameters?.docs?.source}}},p.parameters={...p.parameters,docs:{...p.parameters?.docs,source:{originalSource:`{
  args: {
    loading: true,
    children: "保存中"
  }
}`,...p.parameters?.docs?.source},description:{story:`design-states.md §3 — loading: spinner + aria-busy + disabled、ラベル維持。`,...p.parameters?.docs?.description}}},m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  args: {
    disabled: true,
    children: "保存"
  }
}`,...m.parameters?.docs?.source}}},h.parameters={...h.parameters,docs:{...h.parameters?.docs,source:{originalSource:`{
  render: () => <div style={{
    display: "flex",
    gap: 8
  }}>
      <Button variant="ghost" className="h-9 px-3 text-sm">
        <Trash2 /> 削除
      </Button>
      <Button variant="outline" className="h-9 px-3 text-sm">
        編集
      </Button>
    </div>
}`,...h.parameters?.docs?.source},description:{story:`dense 行アクション例外（design-system.md §7.2: h-9 まで縮小可・min-w-11 維持）。`,...h.parameters?.docs?.description}}}})))()}_();export{d as AllVariants,u as Default,h as DenseRowAction,m as Disabled,p as Loading,f as Sizes,g as __namedExportsOrder,l as default};
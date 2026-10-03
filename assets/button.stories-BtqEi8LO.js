import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{B as n,m as r,t as i}from"./lucide-react-yS7Xowms.js";import{n as a,t as o}from"./button-DzIL5X86.js";var s,c,l,u,d,f,p,m,h;e((()=>{i(),a(),s=t(),c={title:`UI/Button`,component:o,tags:[`autodocs`],argTypes:{variant:{control:`select`,options:[`default`,`primary`,`destructive`,`outline`,`secondary`,`ghost`,`link`,`ghost-danger`]},size:{control:`select`,options:[`default`,`sm`,`lg`,`icon`]},loading:{control:`boolean`},disabled:{control:`boolean`}},args:{children:`保存`},parameters:{docs:{description:{component:`design-system.md §7.2 + design-states.md §1–3 の仕様を全状態で確認するカタログ。
focus-visible ring はキーボード起因のみ発火するため、Storybook 上の検証は
\`button.test.tsx\`（EMR-207 回帰）と axe（addon-a11y）で担保する。`}}}},l={},u={render:()=>(0,s.jsxs)(`div`,{style:{display:`flex`,gap:12,flexWrap:`wrap`,alignItems:`center`},children:[(0,s.jsx)(o,{variant:`default`,children:`primary（既定）`}),(0,s.jsx)(o,{variant:`secondary`,children:`secondary`}),(0,s.jsx)(o,{variant:`outline`,children:`outline`}),(0,s.jsx)(o,{variant:`ghost`,children:`ghost`}),(0,s.jsx)(o,{variant:`link`,children:`link`}),(0,s.jsx)(o,{variant:`destructive`,children:`destructive`}),(0,s.jsx)(o,{variant:`ghost-danger`,children:`ghost-danger`})]})},d={render:()=>(0,s.jsxs)(`div`,{style:{display:`flex`,gap:12,alignItems:`center`},children:[(0,s.jsx)(o,{size:`sm`,children:`sm`}),(0,s.jsx)(o,{size:`default`,children:`default`}),(0,s.jsx)(o,{size:`lg`,children:`lg`}),(0,s.jsx)(o,{size:`icon`,"aria-label":`追加`,children:(0,s.jsx)(n,{})})]})},f={args:{loading:!0,children:`保存中`}},p={args:{disabled:!0,children:`保存`}},m={render:()=>(0,s.jsxs)(`div`,{style:{display:`flex`,gap:8},children:[(0,s.jsxs)(o,{variant:`ghost`,className:`h-9 px-3 text-sm`,children:[(0,s.jsx)(r,{}),` 削除`]}),(0,s.jsx)(o,{variant:`outline`,className:`h-9 px-3 text-sm`,children:`編集`})]})},h=[`Default`,`AllVariants`,`Sizes`,`Loading`,`Disabled`,`DenseRowAction`],l.parameters={...l.parameters,docs:{...l.parameters?.docs,source:{originalSource:`{}`,...l.parameters?.docs?.source},description:{story:`Primary CTA — pill + #038B94。design-system.md §7.2 button-primary。`,...l.parameters?.docs?.description}}},u.parameters={...u.parameters,docs:{...u.parameters?.docs,source:{originalSource:`{
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
}`,...u.parameters?.docs?.source}}},d.parameters={...d.parameters,docs:{...d.parameters?.docs,source:{originalSource:`{
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
}`,...d.parameters?.docs?.source}}},f.parameters={...f.parameters,docs:{...f.parameters?.docs,source:{originalSource:`{
  args: {
    loading: true,
    children: "保存中"
  }
}`,...f.parameters?.docs?.source},description:{story:`design-states.md §3 — loading: spinner + aria-busy + disabled、ラベル維持。`,...f.parameters?.docs?.description}}},p.parameters={...p.parameters,docs:{...p.parameters?.docs,source:{originalSource:`{
  args: {
    disabled: true,
    children: "保存"
  }
}`,...p.parameters?.docs?.source}}},m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
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
}`,...m.parameters?.docs?.source},description:{story:`dense 行アクション例外（design-system.md §7.2: h-9 まで縮小可・min-w-11 維持）。`,...m.parameters?.docs?.description}}}}))();export{u as AllVariants,l as Default,m as DenseRowAction,p as Disabled,f as Loading,d as Sizes,h as __namedExportsOrder,c as default};
import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./calendar-DsXX6KN0.js";var i,a,o,s,c,l;e((()=>{n(),i=t(),a={title:`UI/Calendar`,component:r,tags:[`autodocs`],parameters:{docs:{description:{component:`react-day-picker ラッパー。予約・来院日選択に使う月カレンダー。`}}}},o={render:()=>(0,i.jsx)(r,{mode:`single`,defaultMonth:new Date(2026,8,1)})},s={render:()=>(0,i.jsx)(r,{mode:`single`,selected:new Date(2026,8,15),defaultMonth:new Date(2026,8,1)})},c={render:()=>(0,i.jsx)(r,{mode:`single`,defaultMonth:new Date(2026,8,1),disabled:{dayOfWeek:[3]}})},l=[`Default`,`WithSelection`,`WithDisabledDays`],o.parameters={...o.parameters,docs:{...o.parameters?.docs,source:{originalSource:`{
  render: () => <Calendar mode="single" defaultMonth={new Date(2026, 8, 1)} />
}`,...o.parameters?.docs?.source}}},s.parameters={...s.parameters,docs:{...s.parameters?.docs,source:{originalSource:`{
  render: () => <Calendar mode="single" selected={new Date(2026, 8, 15)} defaultMonth={new Date(2026, 8, 1)} />
}`,...s.parameters?.docs?.source}}},c.parameters={...c.parameters,docs:{...c.parameters?.docs,source:{originalSource:`{
  render: () => <Calendar mode="single" defaultMonth={new Date(2026, 8, 1)} disabled={{
    dayOfWeek: [3]
  }} />
}`,...c.parameters?.docs?.source},description:{story:`診療日など範囲外を選択不可にするパターン（休診日 = 水曜の例）。`,...c.parameters?.docs?.description}}}}))();export{o as Default,c as WithDisabledDays,s as WithSelection,l as __namedExportsOrder,a as default};
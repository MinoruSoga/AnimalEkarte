import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{t}from"./jsx-runtime-DeHZSEgm.js";import{d as n,n as r,s as i}from"./design-tokens-B-kSoC24.js";import{r as a,s as o}from"./table-q93w9vu9.js";function s({staff:e,isValidStaff:t,className:n=``}){return e?t(e)?(0,c.jsx)(a,{className:`${i.tableCell} min-w-[160px] whitespace-nowrap ${n}`,children:(0,c.jsx)(`span`,{className:r.text,children:e})}):u(e)?(0,c.jsx)(a,{className:`${i.tableCell} min-w-[160px] whitespace-nowrap ${n}`,children:(0,c.jsx)(`span`,{className:r.text60,children:`担当者不明`})}):(0,c.jsx)(a,{className:`${i.tableCell} min-w-[160px] whitespace-nowrap ${n}`,children:(0,c.jsxs)(`span`,{className:r.text,children:[e,`（無効）`]})}):(0,c.jsx)(a,{className:`${i.tableCell} min-w-[160px] whitespace-nowrap ${n}`,children:(0,c.jsx)(`span`,{className:r.text60,children:`-`})})}var c,l,u;function d(){return(d=e((()=>{n(),o(),c=t(),l=`担当者不明`,u=e=>e===l||/^\d+$/.test(e),s.__docgenInfo={description:`EMR-412 X3: カルテ一覧・トリミング一覧で共通の担当者セル。
- 空欄は「-」（X15 の記号）
- 名前解決できない生 ID のみは「担当者不明」
- 退職等で無効は「（無効）」を文字で併記（色警告・アイコンに依存しない）
- 名前は C.text。whitespace-nowrap・min-w-[160px]・STYLE.tableCell。`,methods:[],displayName:`StaffNameCell`,props:{staff:{required:!0,tsType:{name:`string`},description:`解決済みの担当者名。doctor_id など名前解決に失敗した生値が入ることもある。`},isValidStaff:{required:!0,tsType:{name:`signature`,type:`function`,raw:`(name: string) => boolean`,signature:{arguments:[{type:{name:`string`},name:`name`}],return:{name:`boolean`}}},description:`アクティブなスタッフ名集合による検証（useStaffValidation）。`},className:{required:!1,tsType:{name:`string`},description:`追加の TableCell クラス（列の表示切替 hidden lg:table-cell など）。`,defaultValue:{value:`""`,computed:!1}}}}})))()}function f({children:e}){return(0,p.jsx)(`table`,{children:(0,p.jsx)(`tbody`,{children:(0,p.jsx)(`tr`,{children:e})})})}var p,m,h,g,_,v,y,b;function x(){return(x=e((()=>{d(),p=t(),m={title:`Shared/StaffNameCell`,component:s,tags:[`autodocs`],parameters:{layout:`padded`}},h=e=>e!==`退職 太郎`,g={render:()=>(0,p.jsx)(f,{children:(0,p.jsx)(s,{staff:``,isValidStaff:h})})},_={render:()=>(0,p.jsx)(f,{children:(0,p.jsx)(s,{staff:`対応 花子`,isValidStaff:h})})},v={render:()=>(0,p.jsx)(f,{children:(0,p.jsx)(s,{staff:`退職 太郎`,isValidStaff:h})})},y={render:()=>(0,p.jsx)(f,{children:(0,p.jsx)(s,{staff:`担当者不明`,isValidStaff:h})})},b=[`Blank`,`ValidStaff`,`InactiveStaff`,`UnresolvedStaff`],g.parameters={...g.parameters,docs:{...g.parameters?.docs,source:{originalSource:`{
  render: () => <TableStory>
      <StaffNameCell staff="" isValidStaff={isValid} />
    </TableStory>
}`,...g.parameters?.docs?.source},description:{story:`空欄。検証せず「-」だけを出す（X15 の記号）。`,...g.parameters?.docs?.description}}},_.parameters={..._.parameters,docs:{..._.parameters?.docs,source:{originalSource:`{
  render: () => <TableStory>
      <StaffNameCell staff="対応 花子" isValidStaff={isValid} />
    </TableStory>
}`,..._.parameters?.docs?.source},description:{story:`アクティブな担当者。氏名は ink（C.text）。`,..._.parameters?.docs?.description}}},v.parameters={...v.parameters,docs:{...v.parameters?.docs,source:{originalSource:`{
  render: () => <TableStory>
      <StaffNameCell staff="退職 太郎" isValidStaff={isValid} />
    </TableStory>
}`,...v.parameters?.docs?.source},description:{story:`退職等で無効になった担当者。「（無効）」を文字で併記し色警告に依存しない。`,...v.parameters?.docs?.description}}},y.parameters={...y.parameters,docs:{...y.parameters?.docs,source:{originalSource:`{
  render: () => <TableStory>
      <StaffNameCell staff="担当者不明" isValidStaff={isValid} />
    </TableStory>
}`,...y.parameters?.docs?.source},description:{story:`名前解決できない担当者（doctor_id のみ等）。「担当者不明」。`,...y.parameters?.docs?.description}}}})))()}x();export{g as Blank,v as InactiveStaff,y as UnresolvedStaff,_ as ValidStaff,b as __namedExportsOrder,m as default};
import{n as e}from"./rolldown-runtime-DkW27tQK.js";import{n as t,t as n}from"./BirthdayBadge-RE4Z7yo6.js";var r,i,a,o,s,c,l,u,d;function f(){return(f=e((()=>{t(),r=new Date(Date.now()+324e5),i=e=>`${e.getUTCFullYear()}-${String(e.getUTCMonth()+1).padStart(2,`0`)}-${String(e.getUTCDate()).padStart(2,`0`)}`,a=i(new Date(Date.UTC(r.getUTCFullYear(),r.getUTCMonth(),15))),o=i(new Date(Date.UTC(r.getUTCFullYear(),r.getUTCMonth()-1,15))),s={title:`Shared/BirthdayBadge`,component:n,tags:[`autodocs`],args:{subjectName:`ポチ`,birthDate:a}},c={},l={args:{variant:`static`,stopPropagation:!0}},u={args:{birthDate:o}},d=[`Default`,`Static`,`NotBirthMonth`],c.parameters={...c.parameters,docs:{...c.parameters?.docs,source:{originalSource:`{}`,...c.parameters?.docs?.source}}},l.parameters={...l.parameters,docs:{...l.parameters?.docs,source:{originalSource:`{
  args: {
    variant: "static",
    stopPropagation: true
  }
}`,...l.parameters?.docs?.source}}},u.parameters={...u.parameters,docs:{...u.parameters?.docs,source:{originalSource:`{
  args: {
    birthDate: prevMonth
  }
}`,...u.parameters?.docs?.source}}}})))()}f();export{c as Default,u as NotBirthMonth,l as Static,d as __namedExportsOrder,s as default};
import{i as e}from"./preload-helper-xPQekRTU.js";import{t}from"./jsx-runtime-CaZkqeYb.js";import{n,t as r}from"./badge-DxN69KAE.js";import{a as i,i as a,n as o,o as s,r as c,s as l,t as u}from"./table-tkYgzSSj.js";var d,f,p,m,h;e((()=>{l(),n(),d=t(),f={title:`UI/Table`,component:u,tags:[`autodocs`]},p=[{time:`09:30`,pet:`モモ`,owner:`山田`,status:`診察中`},{time:`10:00`,pet:`チョコ`,owner:`佐藤`,status:`待合`},{time:`10:30`,pet:`リン`,owner:`鈴木`,status:`会計待ち`}],m={render:()=>(0,d.jsxs)(u,{children:[(0,d.jsx)(i,{children:(0,d.jsxs)(s,{children:[(0,d.jsx)(a,{children:`時刻`}),(0,d.jsx)(a,{children:`ペット`}),(0,d.jsx)(a,{children:`飼主`}),(0,d.jsx)(a,{children:`状態`})]})}),(0,d.jsx)(o,{children:p.map(e=>(0,d.jsxs)(s,{children:[(0,d.jsx)(c,{children:e.time}),(0,d.jsx)(c,{children:e.pet}),(0,d.jsx)(c,{children:e.owner}),(0,d.jsx)(c,{children:(0,d.jsx)(r,{variant:`secondary`,children:e.status})})]},e.time))})]})},h=[`Default`],m.parameters={...m.parameters,docs:{...m.parameters?.docs,source:{originalSource:`{
  render: () => <Table>
      <TableHeader>
        <TableRow>
          <TableHead>時刻</TableHead>
          <TableHead>ペット</TableHead>
          <TableHead>飼主</TableHead>
          <TableHead>状態</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {ROWS.map(r => <TableRow key={r.time}>
            <TableCell>{r.time}</TableCell>
            <TableCell>{r.pet}</TableCell>
            <TableCell>{r.owner}</TableCell>
            <TableCell>
              <Badge variant="secondary">{r.status}</Badge>
            </TableCell>
          </TableRow>)}
      </TableBody>
    </Table>
}`,...m.parameters?.docs?.source}}}}))();export{m as Default,h as __namedExportsOrder,f as default};
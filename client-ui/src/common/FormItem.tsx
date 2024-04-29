import * as React from 'react';
import {Form} from "antd"

type FormItemProps = {
    name: string
    label: string
    children: React.ReactNode
}

export function FormItem({name, label, children}: FormItemProps) {
    return <Form.Item
        name={name}
        label={label}
        labelAlign={'left'}
        labelCol={{span: 10}}
        wrapperCol={{span: 14}}
        colon={false}
    >
        {children}
    </Form.Item>
}
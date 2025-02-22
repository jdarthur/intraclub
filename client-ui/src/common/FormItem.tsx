import * as React from 'react';
import {DatePicker, Form, Input, InputNumber, Select, TimePicker} from "antd"
import dayjs from "dayjs";

type FormItemProps = {
    name: string
    label: string
    children?: React.ReactNode
    inputType?: string
    disabled?: boolean
    placeholder?: string
}

const INPUT = "input"
const SUGGESTION_INPUT = "suggestion_input"
const NUMBER_INPUT = "number_input"
const TEXT_AREA = "text_area"

export function FormItem({name, label, children, inputType, disabled, placeholder}: FormItemProps) {

    let content = children
    if (!content && inputType) {

        if (inputType == INPUT) {
            content = <Input disabled={disabled} placeholder={placeholder}/>
        } else if (inputType == NUMBER_INPUT) {
            content = <InputNumber disabled={disabled} placeholder={placeholder}/>
        } else if (inputType == TEXT_AREA) {
            content = <Input.TextArea disabled={disabled} placeholder={placeholder}/>
        } else if (inputType == SUGGESTION_INPUT) {
            content = <Input disabled={disabled} placeholder={placeholder}/>
        }
    }

    return <Form.Item
        name={name}
        label={label}
        labelAlign={'left'}
        labelCol={{span: 10}}
        wrapperCol={{span: 14}}
        colon={false}
    >
        {content}
    </Form.Item>
}

export type NameAndLabel = {
    name: string
    label: string
    disabled?: boolean
    placeholder?: string
}

export function InputFormItem({name, label, disabled, placeholder}: NameAndLabel) {
    return <FormItem name={name} label={label} inputType={INPUT} disabled={disabled} placeholder={placeholder}/>
}

export function NumberInputFormItem({name, label}: NameAndLabel) {
    return <FormItem name={name} label={label} inputType={NUMBER_INPUT}/>
}


export function TextAreaFormItem({name, label}: NameAndLabel) {
    return <FormItem name={name} label={label} inputType={TEXT_AREA}/>
}

type SelectProps = NameAndLabel & {
    options: { label: string, value: string }[]
}

export function SelectFormItem({name, label, options, disabled}: SelectProps) {
    return <FormItem name={name} label={label}>
        <Select options={options} disabled={disabled}/>
    </FormItem>
}

export function TimePickerFormItem({name, label, disabled}: NameAndLabel) {
    return <FormItem name={name} label={label}>
        <TimePicker use12Hours minuteStep={15} format={"HH:mm"}
                    showNow={false} needConfirm={false} disabled={disabled}/>
    </FormItem>
}

type DatePickerProps = NameAndLabel & {
    multiple?: boolean
    future?: boolean
}

export function DatePickerFormItem({name, label, multiple, future, disabled}: DatePickerProps) {
    return <FormItem name={name} label={label}>
        <DatePicker multiple={multiple} minDate={future ? dayjs() : null} format={"YYYY-MM-DD"} disabled={disabled}/>
    </FormItem>
}


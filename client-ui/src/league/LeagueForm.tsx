import * as React from 'react';
import {useToken} from "../redux/auth.js";
import {Button, DatePicker, Form, Input, Modal, Select, TimePicker} from "antd";
import {PlusSquareOutlined} from "@ant-design/icons";
import {TeamColorProps} from "../team/TeamColor";
import dayjs from 'dayjs'
import {useWhoAmIQuery} from "../redux/api";
import {Facility} from "../settings/Facilities";

export type League = {
    league_id?: string
    name: string
    colors?: TeamColorProps[]
    commissioner: string
    reporters?: string[]
    facility: string
    start_time: number
    weeks: string[]
    active?: boolean
}

type Week = {
    week_id?: string
    date: string
    original_date: string
}

type LeagueFormProps = {
    Update?: boolean
    InitialState?: League
}

export function LeagueForm() {

    const [open, setOpen] = React.useState<boolean>(false);

    const [form] = Form.useForm()

    const {data} = useWhoAmIQuery()

    const token = useToken()

    const onSave = (v: any) => {
        const formValues = form.getFieldsValue();
        console.log(formValues, v)

        const weeks: string[] = []
        for (let i = 0; i < formValues.weeks.length; i++) {
            const w = formValues.weeks[i]?.format("YYYY-MM-DD")
            const week = CreateWeek(w)
            weeks.push(week.week_id)
        }

        const body: League = {
            name: formValues.name,
            commissioner: data?.user_id,
            weeks: weeks,
            start_time: formValues.start_time?.format("HH:mm"),
            facility: formValues.facility,
        }

        console.log(body)
    }

    return <div>
        <Modal open={open} title={"Create a new league"} onCancel={() => setOpen(false)}>
            <Form form={form}
                  onFinish={onSave}>
                <Form.Item name={"name"} label={"Name"}>
                    <Input placeholder={"League name"}/>
                </Form.Item>
                <Form.Item name={"facility"} label={"Facility"}>
                    <Select options={[{label: "Facility 1", value: "1"}, {label: "Facility 2", value: "2"}]}/>
                </Form.Item>
                <Form.Item name={"start_time"} label={"Start time"}>
                    <TimePicker use12Hours minuteStep={15} format={"HH:mm"}
                                showNow={false} needConfirm={false}/>
                </Form.Item>
                <Form.Item name={"weeks"} label={"Weeks"}>
                    <DatePicker multiple minDate={dayjs()}/>
                </Form.Item>
                <Form.Item>
                    <Button type="primary" htmlType="submit">
                        Submit
                    </Button>
                </Form.Item>
            </Form>
        </Modal>
        {token ? <Button style={{marginTop: "2em"}} type={"primary"} onClick={() => setOpen(true)}>
            <PlusSquareOutlined style={{marginRight: "0.5em"}}/>
            Create a new league
        </Button> : null}
    </div>
}


function CreateWeek(date: string): Week {
    return {
        week_id: "test",
        date: date,
        original_date: date,
    }
}

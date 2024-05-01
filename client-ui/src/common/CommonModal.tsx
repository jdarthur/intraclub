import * as React from 'react';
import {Alert, Button, Modal} from "antd";
import {EditOutlined, PlusSquareOutlined} from "@ant-design/icons";
import {onOffline} from "@reduxjs/toolkit/dist/query/core/setupListeners";


type CommonModalProps = {
    ObjectType: string
    IsUpdate: boolean
    OnSubmit: () => Promise<SubmitResult>
    children: React.ReactNode
}

export type SubmitResult = {
    data: {}
    error: { data: { error: string } }
}

export function CommonModal({ObjectType, IsUpdate, OnSubmit, children}: CommonModalProps) {
    const [open, setOpen] = React.useState(false);
    const [error, setError] = React.useState<string>("")

    const title = IsUpdate ? `Update ${ObjectType}` : `Create a new ${ObjectType}`
    const okText = IsUpdate ? "Update" : "Create"

    // default button to open the modal
    const DefaultButton = <Button type={"primary"} icon={<PlusSquareOutlined/>}>
        New {ObjectType}
    </Button>

    // on update, we will just show an edit icon
    const actionButton = <div onClick={() => setOpen(true)}>
        {IsUpdate ? <EditOutlined style={{cursor: "pointer"}}/> : DefaultButton}
    </div>

    const onOk = async () => {
        const result = await OnSubmit()

        if (result.data) {
            setOpen(false)
            setError("")
        } else {
            setError(result.error.data.error)
        }
    }


    return <div>
        <Modal open={open} title={title} onCancel={() => setOpen(false)} onOk={onOk} okText={okText}>
            <div style={{height: "1em"}}/>
            {children}

            {/* show an error alert if the OnSubmit function returned us an error */}
            {error ? <Alert type={"error"} description={error} style={{padding: "0.7em"}}/> : null}
        </Modal>
        {actionButton}
    </div>
}
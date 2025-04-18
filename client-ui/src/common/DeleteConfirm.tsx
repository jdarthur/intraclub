import * as React from 'react';
import {Button, notification, Popover} from "antd";
import {DeleteColumnOutlined, DeleteOutlined} from "@ant-design/icons";

type DeleteConfirmProps = {
    deleteFunction: () => string
    objectType: string
}

export function DeleteConfirm({deleteFunction, objectType}: DeleteConfirmProps) {
    const [open, setOpen] = React.useState<boolean>(false);


    const [api] = notification.useNotification();

    const errorNotification = (title: string, error: string) => {
        api.error({
            message: title,
            description: error,
        });
    };

    const onDelete = () => {
        const error = deleteFunction();
        if (error) {
            setOpen(false);
            errorNotification(`Error deleting ${objectType}`, error);
        }
    }

    const content = <div style={{display: "flex", justifyContent: "flex-end"}}>
        <Button onClick={() => setOpen(false)} size={"small"} style={{marginRight: "0.3em"}}>
            Cancel
        </Button>
        <Button onClick={onDelete} danger type={"primary"} size={"small"}>
            Delete
        </Button>
    </div>


    return <Popover open={open} title={`Delete ${objectType}?`} content={content}>
        <DeleteOutlined onClick={() => setOpen(true)} style={{cursor: 'pointer'}}/>
    </Popover>
}
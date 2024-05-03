import * as React from 'react';
import {Alert, Button, Form, Modal} from "antd";
import {EditOutlined, PlusSquareOutlined} from "@ant-design/icons";

type CommonModalProps = {
    // name of the record, e.g. `"facility"`, `"league"`, etc.
    //  - this is used to populate the text in the create button / modal title
    ObjectType: string

    // `true` if this is an update call
    //   - used for submit button text / part of the modal's title)
    IsUpdate: boolean

    // this is called when you hit the submit button on the form
    //  - the `formValues` prop passed in will be the result from `form.getFieldsValue()`
    //  and the function is expected to asynchronously return a `SubmitResult` value.
    OnSubmit: (formValues: any) => Promise<SubmitResult>

    // children of the form itself, e.g. a bunch of `FormItem`s
    children: React.ReactNode

    // the initial form state, used to pre-populate the form on an update
    InitialState: {}
}

// SubmitResult is the expected result from an `OnSubmit` call in `CommonModalProps`
// This basically just wraps around the redux mutation API response structure, so you
// can return the value of `await myMutationCall(args)`
export type SubmitResult = {
    data: {}
    error: { data: { error: string } }
}

// CommonFormModal is a component that wraps the <Modal> component from antd with a
// <Form> component inside. This component does some common styling / formatting, e.g.
// consistent values for the modal title, submit button text, etc.
//
// It also wraps the <Form> component and its state inside of this object so you don't
// have to care about it in the component implementing any specific form.
//
// The open/closed state of the modal is handled in internal state here and the component
// will render when closed either as an `<EditOutlined />` component or a `<Button />`
// that has text of the format `Create a new [ObjectType]`
//
// The onOk function will parse the success/failure state from the `SubmitResult` it receives
// and either close the modal on success or populate an error `<Alert/>` with the particular error
// message that we received from the API
export function CommonFormModal({ObjectType, IsUpdate, OnSubmit, children, InitialState}: CommonModalProps) {
    // controls the open/closed state of the modal
    const [open, setOpen] = React.useState(false);

    // set to a value when we hit the submit button and receive an error from the API
    const [error, setError] = React.useState<string>("")

    // handle the state of the form with this object
    const [form] = Form.useForm()

    // title of the modal
    const title = IsUpdate ? `Update ${ObjectType}` : `Create a new ${ObjectType}`

    // title of the submit button at the bottom of the modal
    const okText = IsUpdate ? "Update" : "Create"

    // default button to open the modal when IsUpdate == false
    const DefaultButton = <Button type={"primary"} icon={<PlusSquareOutlined/>}>
        Create a new {ObjectType}
    </Button>

    // on update, we will just show an edit icon
    const actionButton = <div onClick={() => setOpen(true)}>
        {IsUpdate ? <EditOutlined style={{cursor: "pointer"}}/> : DefaultButton}
    </div>

    // this function calls the OnSubmit function passed as an argument, checks if we
    // got a
    const onOk = async () => {
        const formValues = form.getFieldsValue()
        const result = await OnSubmit(formValues)

        if (result.data) {
            setOpen(false)
            setError("")
        } else {
            setError(result.error.data.error)
        }
    }

    const printChanges = (v: any) => {
        //console.log(v)
    }

    return <div>
        <Modal open={open} title={title} onCancel={() => setOpen(false)} onOk={onOk} okText={okText}>
            <div style={{height: "1em"}}/>
            <Form form={form} layout={"horizontal"} initialValues={InitialState} onValuesChange={printChanges}>
                {children}
            </Form>

            {/* show an error alert if the OnSubmit function returned us an error */}
            {error ? <Alert type={"error"} description={error} style={{padding: "0.7em"}}/> : null}
        </Modal>
        {actionButton}
    </div>
}
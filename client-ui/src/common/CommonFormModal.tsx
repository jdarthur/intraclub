import * as React from 'react';
import {Alert, Button, Form, FormInstance, Modal} from "antd";
import {EditOutlined, PlusSquareOutlined} from "@ant-design/icons";

export type CommonModalProps = {
    // name of the record, e.g. `"facility"`, `"league"`, etc.
    //  - this is used to populate the text in the create button / modal title
    ObjectType: string

    // Custom title for the button that opens the modal
    title?: string

    // `true` if this is an update call
    //   - used for submit button text / part of the modal's title)
    IsUpdate: boolean

    // this is called when you hit the submit button on the form
    //  - the `formValues` prop passed in will be the result from `form.getFieldsValue()`
    //  and the function is expected to asynchronously return a `SubmitResult` value.
    //
    // this can be set to null if you have a different mechanism for submitting the form,
    // e.g. via a step form or something like that
    OnSubmit?: (formValues: any) => Promise<SubmitResult>

    // children of the form itself, e.g. a bunch of `FormItem`s
    children: React.ReactNode

    // the initial form state, used to pre-populate the form on an update
    InitialState: {}

    // allows you to pass a form instance into the form instead of using a local one
    form?: FormInstance

    // allows you to set the footer manually including to a {null} value
    footer?: any

    open?: boolean
    setOpen?: (b: boolean) => void
    onCancel?: () => Promise<SubmitResult>
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
export function CommonFormModal({...props}: CommonModalProps) {
    // controls the open/closed state of the modal
    const [localOpen, setLocalOpen] = React.useState(false);

    // set to a value when we hit the submit button and receive an error from the API
    const [error, setError] = React.useState<string>("")

    let formToUse = props.form
    if (!props.form) {
        // handle the state of the form with this object
        const [f] = Form.useForm()
        formToUse = f
    }

    let realOpen = localOpen
    if (props.open != undefined) {
        realOpen = props.open
    }

    const setOpenWrapper = (b: boolean) => {
        if (props.setOpen) {
            props.setOpen(b)
        } else {
            setLocalOpen(b)
        }
    }


    // title of the modal
    let title = props.IsUpdate ? `Update ${props.ObjectType}` : `Create a new ${props.ObjectType}`
    if (props.title != "") {
        title = props.title;
    }

    // title of the submit button at the bottom of the modal
    const okText = props.IsUpdate ? "Update" : "Create"

    // default button to open the modal when IsUpdate == false
    const DefaultButton = <Button type={"primary"} icon={<PlusSquareOutlined/>}>
        {title}
    </Button>

    // on update, we will just show an edit icon
    const actionButton = <div onClick={() => setOpenWrapper(true)}>
        {props.IsUpdate ? <EditOutlined style={{cursor: "pointer"}}/> : DefaultButton}
    </div>

    // this function calls the OnSubmit function passed as an argument, checks if we
    // got a
    const onOk = async () => {
        const formValues = formToUse.getFieldsValue()
        const result = await props.OnSubmit(formValues)

        if (result.data) {
            setLocalOpen(false)
            setError("")
        } else {
            setError(result.error.data.error)
        }
    }

    const printChanges = (v: any) => {
        //console.log(v)
    }

    const clickCancel = async (): Promise<SubmitResult> => {
        if (props.onCancel) {
            const res = await props.onCancel()
            console.log(res)
            if (res.error) {
                console.log(res.error)
                return res
            }
        }
        setOpenWrapper(false)
        return {data: null, error: null}
    }

    return <div>
        <Modal open={realOpen} title={title} onCancel={clickCancel} onOk={onOk} okText={okText}
               footer={props.footer}>
            <div style={{height: "1em"}}/>
            <Form form={formToUse} layout={"horizontal"} initialValues={props.InitialState}
                  onValuesChange={printChanges}>
                {props.children}
            </Form>

            {/* show an error alert if the OnSubmit function returned us an error */}
            {error ? <Alert type={"error"} description={error} style={{padding: "0.7em"}}/> : null}
        </Modal>
        {actionButton}
    </div>
}
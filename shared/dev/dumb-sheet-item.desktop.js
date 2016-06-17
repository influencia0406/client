// @flow
import React, {Component} from 'react'
import ReactDOM from 'react-dom'
import {Box, Text} from '../common-adapters'
import {globalStyles} from '../styles/style-guide'

class DumbSheetItem extends Component<void, Props, void> {
  componentDidMount () {
    if (this.props.mock.afterMount) {
      this.props.mock.afterMount(this._component, ReactDOM.findDOMNode(this._component))
    }
  }

  render () {
    const Component = this.props.component
    const {parentProps, afterMount, ...mock} = this.props.mock
    return (
      <Box id={this.props.id} style={{...styleBox, ...this.props.style}}>
        <Text type='Body' style={{marginBottom: 5}}>{this.props.mockKey}</Text>
        <Box {...parentProps}>
          <Component ref={c => this._component = c} {...mock} />
        </Box>
      </Box>
    )
  }
}

const styleBox = {
  ...globalStyles.flexBoxColumn,
  padding: 20,
  marginTop: 10,
  border: 'solid 1px lightgray',
  boxShadow: '5px 5px lightgray'
}

export default DumbSheetItem

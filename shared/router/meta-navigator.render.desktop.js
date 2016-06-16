import React, {Component} from 'react'
import {globalStyles} from '../styles/style-guide'

class MetaNavigatorRender extends Component {
  render () {
    const {rootComponent, uri, getComponentAtTop} = this.props
    const {componentAtTop} = getComponentAtTop(rootComponent, uri)
    const Module = componentAtTop.component
    const element = componentAtTop.element

    return (
      <div style={{...globalStyles.flexBoxColumn, flex: 1}}>
        {element}
        {!element && Module && <Module {...componentAtTop.props} />}
      </div>
    )
  }
}

export default MetaNavigatorRender

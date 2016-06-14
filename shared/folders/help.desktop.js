// @flow
import React from 'react'
import {Box, Text, Terminal} from '../common-adapters'
import {globalStyles, globalColors} from '../styles/style-guide'
import {shell} from 'electron'

const RenderHelp = () => (
  <Box style={stylesScrollContainer}>
    <Box style={stylesContainer}>
      <Box style={styleIconHeader}>
        // TODO (AW): add header icon
      </Box>
      <Box style={styleTextHeader}>
        <Text type='Body'>
          Soon, this is where you'll manage and learn about folders in the Keybase
          filesystem. It will be <Text
            type='BodyPrimaryLink'
            onClick={() => shell.openExternal('https://www.youtube.com/watch?v=kTOvzgNsOXw')}>
            incredibly powerful
          </Text> when finished.
        </Text>
      </Box>
      <Box style={styleBody}>
        <Text type='BodySmall' style={{...styleBodyText}}>
          Here are a few terminal examples for the meantime. Note you can share with
          people who haven't joined Keybase yet, and your computer will rekey the
          data for them the moment they establish keys.
        </Text>
        <Text type='BodySmall' style={{...styleBodyText, ...globalStyles.italic}}>
          True end-to-end crypto.
        </Text>
        <Terminal style={styleTerminal}>
          <Text type='Terminal'>cd /keybase/public/cnojima</Text>
          <Text type='Terminal'>cd /keybase/private/cnojima</Text>
          <Text type='Terminal'>cd /keybase/private/cnojima,chris</Text>
          <Text type='TerminalEmpty' />
          <Text type='TerminalComment'>works even before maxtaco has joined keybase:</Text>
          <Text type='Terminal'>cd /keybase/private/cnojima,maxtaco@twitter</Text>
          <Text type='TerminalEmpty' />
          <Text type='TerminalComment'>OSX tip: this opens Finder</Text>
          <Text type='Terminal'>open /keybase/private/cnojima</Text>
          <Text type='TerminalEmpty' />
          <Text type='TerminalComment'>to hide a folder in your Finder</Text>
          <Text type='Terminal'>rmdir /keybase/private/cnojima,some_enemy</Text>
        </Terminal>
      </Box>
    </Box>
  </Box>
)

const stylesScrollContainer = {
  ...globalStyles.scrollable,
  flexGrow: 1,
  background: globalColors.lightGrey
}

const stylesContainer = {
  ...globalStyles.flexBoxColumn,
  alignItems: 'center',
  textAlign: 'center',
  background: globalColors.white
}

const styleIconHeader = {
  marginTop: 64,
  height: 80,
  marginBottom: 16
}

const styleTextHeader = {
  maxWidth: 512,
  marginLeft: 4,
  marginRight: 4,
  marginBottom: 32
}

const styleBody = {
  ...globalStyles.flexBoxColumn,
  alignItems: 'center',
  textAlign: 'center',
  flexGrow: 1,
  background: globalColors.lightGrey,
  padding: 32,
  width: '100%'
}

const styleBodyText = {
  maxWidth: 512
}

const styleTerminal = {
  borderRadius: 4,
  padding: 32,
  marginTop: 16,
  width: '100%',
  maxWidth: 576
}

export default RenderHelp

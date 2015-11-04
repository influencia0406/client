'use strict'
/* @flow */

import React, { Component, StyleSheet, View, Text } from 'react-native'
import { showedSecretWords } from '../actions/login'
import commonStyles from '../styles/common'

export default class DisplaySecretWords extends Component {
  constructor (props) {
    super(props)

    props.onSubmit()
  }

  render () {
    return (
        <View style={styles.container}>
          <Text style={[{textAlign: 'center', marginBottom: 75}, commonStyles.h1]}>Register Device</Text>
          <Text style={[{margin: 20, marginBottom: 20}, commonStyles.h2]}>In order to register this device you need to enter in the secret phrase generated on an existing device</Text>
          <Text style={[styles.secret, commonStyles.h1]}>{this.props.secretWords}</Text>
        </View>
    )
  }

  static parseRoute (store) {
    const { secretWords, response } = store.getState().login

    return {
      componentAtTop: {
        title: 'Register Device',
        leftButtonTitle: 'Cancel',
        mapStateToProps: state => state.login,
        props: {
          onSubmit: () => store.dispatch(showedSecretWords(response)),
          secretWords
        }
      }
    }
  }
}

DisplaySecretWords.propTypes = {
  response: React.PropTypes.object,
  secretWords: React.PropTypes.string
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'stretch',
    backgroundColor: '#F5FCFF'
  },
  secret: {
    textAlign: 'center',
    marginBottom: 75,
    backgroundColor: 'grey',
    borderColor: 'black',
    padding: 10
  }
})
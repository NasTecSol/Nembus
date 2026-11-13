//
//  TimerSoundService.swift
//  NembusWidget
//

import Foundation
import AVFoundation
import AudioToolbox

final class TimerSoundService: NSObject, AVAudioPlayerDelegate {
    static let shared = TimerSoundService()

    private var audioPlayer: AVAudioPlayer?
    private var registeredSoundID: SystemSoundID = 0

    override init() {
        super.init()
        setupAudio()
    }

    func setupAudio() {
        if let url = findAudioURL() {
            AudioServicesCreateSystemSoundID(url as CFURL, &registeredSoundID)
        }
    }

    func findAudioURL() -> URL? {
        let fileNames = [
            ("timer", "caf"),
            ("timer", "wav"),
            ("timer", "mp3")
        ]

        for (name, ext) in fileNames {
            if let url = Bundle.main.url(forResource: name, withExtension: ext) {
                return url
            }
            if let url = Bundle.main.url(forResource: name, withExtension: ext, subdirectory: "assets") {
                return url
            }
            if let url = Bundle.main.url(forResource: "flutter_assets/assets/\(name)", withExtension: ext) {
                return url
            }
            if let libDir = FileManager.default.urls(for: .libraryDirectory, in: .userDomainMask).first {
                let soundPath = libDir.appendingPathComponent("Sounds/\(name).\(ext)")
                if FileManager.default.fileExists(atPath: soundPath.path) {
                    return soundPath
                }
            }
        }
        return nil
    }

    func playTimerSound() {
        if let soundURL = findAudioURL() {
            do {
                let session = AVAudioSession.sharedInstance()
                try session.setCategory(.playback, mode: .default, options: [.duckOthers])
                try session.setActive(true)
                audioPlayer = try AVAudioPlayer(contentsOf: soundURL)
                audioPlayer?.delegate = self
                audioPlayer?.prepareToPlay()
                if audioPlayer?.play() == true {
                    return
                }
            } catch {
                print("AVAudioPlayer error: \(error.localizedDescription)")
            }
        }

        if registeredSoundID != 0 {
            AudioServicesPlayAlertSound(registeredSoundID)
        } else {
            AudioServicesPlayAlertSound(SystemSoundID(1304))
            AudioServicesPlayAlertSound(SystemSoundID(kSystemSoundID_Vibrate))
        }
    }
}

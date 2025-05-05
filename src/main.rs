use aes::{
    Aes256,
    cipher::{Block, BlockEncrypt, KeyInit, generic_array::GenericArray},
};
use flate2::read::MultiGzDecoder;
use std::fs::File;
use std::io::{BufReader, Write};
use vcf::{VCFError, VCFReader, VCFRecord};

fn encrypt_vcf_data(key: &[u8], data: &[u8]) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
    let cipher = Aes256::new_from_slice(key).map_err(|_| "Invalid key length")?;

    // Pad data to multiple of 16 bytes (AES block size)
    let block_size = 16;
    let padded_len = ((data.len() + block_size - 1) / block_size) * block_size;
    let mut padded_data = vec![0u8; padded_len];
    padded_data[..data.len()].copy_from_slice(data);

    // Encrypt each 16-byte block
    let mut encrypted = Vec::new();
    for chunk in padded_data.chunks(block_size) {
        let mut block = [0u8; 16];
        block.copy_from_slice(chunk);
        cipher.encrypt_block(GenericArray::from_mut_slice(&mut block));
        encrypted.extend_from_slice(&block);
    }

    Ok(encrypted)
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 32-byte key for AES-256
    let key = hex::decode("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
        .map_err(|_| "Invalid key format")?;

    // Read VCF file
    let mut reader = VCFReader::new(BufReader::new(MultiGzDecoder::new(File::open(
        "./data/ALL.chr22.shapeit2_integrated_snvindels_v2a_27022019.GRCh38.phased.vcf.gz",
    )?)))?;

    let mut encrypted_records = Vec::new();
    let mut vcf_record = reader.empty_record();

    // Read and encrypt each record
    let mut record_count = 0;
    while reader.next_record(&mut vcf_record)? {
        let record_data = format!("{:?}", vcf_record).into_bytes();
        let encrypted = encrypt_vcf_data(&key, &record_data)?;
        encrypted_records.push(encrypted);

        record_count += 1;
        if record_count % 100 == 0 {
            println!("Processed {} records", record_count);
        }
    }
    println!("Finished processing a total of {} records", record_count);

    // Save encrypted records to a file
    let mut output_file = File::create("encrypted_vcf_records.bin")?;
    for encrypted in encrypted_records {
        output_file.write_all(&encrypted)?;
    }

    Ok(())
}
